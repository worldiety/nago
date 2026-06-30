# Nago Event hub

Ziel des Nago Event Hubs ist es, Teams und ihre gegenseitigen Abhängigkeiten zu dokumentieren, sowie ihre Verantwortlichkeiten innerhalb einer Unternehmensarchitektur mittels Zugriffsregeln auch durchzusetzen.

Dazu stellt der Event Hub auch einen Message Broker zur Verfügung, über den verschiedene Prozesse via publish-subscribe Mechanik kommunizieren können.

## Topics

Topics dienen üblicherweise einer fachlichen Gliederung, die vorab bekannt sein muss.
Kommen neue Erkenntnisse hinzu oder ändert sich die Fachlichkeit, müssen die Topic-Spezifikation wie Schemas, ACL oder Retention-Regeln ebenfalls angepasst werden.
Wenn dies zu Inkompatibilitäten führen würde, wären mitunter versionierte Topics sowie Event-Duplikation erforderlich.

Dies lässt sich durch den Verzicht auf Topics im Generellen einfach auflösen.
Systeme wie Kafka oder MQTT et al. verwenden Topics vor allem zur Optimierung der Performance. Hiermit verbunden sind bei Kafka insbesondere die Partitionen (jedes Topic hat mind. 1 Partition) und den damit verbundenen Partitionierungsschlüssel.

Wir gehen davon aus, dass es andere Kriterien jenseits von Topic und Partition gibt, um übliche und ausreichende Performanceoptimierungen automatisch anzuwenden.

## Events

Ereignisse (Event) bzw. Nachrichten (Message) haben einen genauen Typ, der dem Schema-Eintrag entsprechen muss.

### Betriebsmodi: Historie (Append) vs. Retained Value (Put)

Der Event Store unterstützt zwei Schreib-Semantiken pro Event-Typ:

#### Append / Replay – Event-Sourcing mit vollständiger Historie

`Append` hängt ein neues Event an die Segment-Datei des jeweiligen Typs an.
Alle Events werden in chronologischer Reihenfolge persistiert und sind via `Replay` mit Sequenz-ID-Bereich abrufbar.
Der Time-Index wird aktualisiert, Segment-Splitting wird angewendet.
Dies ist der klassische Event-Sourcing-Modus.

#### Put / Get – MQTT-Retain-Semantik (letzter bekannter Wert)

`Put` überschreibt immer den einzigen Wert im Pending-Segment des Typs.
Es wird keine Historie aufgebaut – nur der letzte bekannte Wert ist physisch vorhanden.
Dies entspricht der Retained-Message-Semantik von MQTT: ein Subscriber erhält beim Verbinden immer den aktuellen Zustand.

`Get` liest den letzten geschriebenen Wert für einen Event-Typ in O(1) direkt über einen gecachten Datei-Offset – ohne Segment-Scan.
`Get` funktioniert gleichermaßen effizient für beide Modi: bei History-Typen wird das zuletzt per `Append` geschriebene Event zurückgegeben, bei Retain-Typen das per `Put` gesetzte.

Folgende Eigenschaften gelten für Put:

* **Kein Time-Index**: Da keine Historie existiert, wird der Zeitindex nicht aktualisiert.
* **Kein Segment-Splitting**: Das Pending-Segment enthält immer genau eine Message.
* **Globale Sequenz-ID**: Die Sequenz-ID wird wie bei Append vom globalen monotonen Zähler vergeben, sodass Konsumenten Änderungen über Sequenzvergleich erkennen können.
* **Kein Truncate erforderlich**: Wenn ein Put einen kleineren Payload über einen größeren schreibt, verbleiben Stale-Bytes am Ende der Segment-Datei. Diese sind harmlos, da sie mitten im alten Frame beginnen und daher keinen gültigen SyncMarker + CRC bilden können. `readMessages` und `Replay` stoppen an der ungültigen Grenze. Beim nächsten DB-Open bereinigt `repairTailTruncation` den Tail einmalig. `Get` liest direkt über den bekannten Offset und die im Frame kodierte `TotalLen` und sieht den Tail nie.
* **Replay-Kompatibilität**: `Replay` auf einem Put-Typ liefert transparent genau eine Message – die aktuelle. Es ist kein Sonderverhalten implementiert oder erforderlich.
* **Mischbetrieb**: Ob ein Typ per Append oder Put beschrieben wird, wird von der aufrufenden Schicht festgelegt. Der Store selbst erzwingt keine Trennung. Selbst bei versehentlichem Mischbetrieb entsteht kein korruptes Verhalten – Put überschreibt lediglich den ersten Slot, während Append dahinter anhängt.

## Performance

Obwohl das System über eine Schemaregistry verfügt, wird das Schema nicht validiert.
Hintergrund ist in erster Linie die damit verbundene Performanceeinschränkung, da die Implementierung so die Möglichkeit hat, die Events ohne Konvertierung direkt auszuliefern und insbesondere die dafür erforderliche CPU und Speicherbandbreite zu minimieren.

Die Metadaten Sequenz-ID, Zeitstempel und Tracing-ID sind gesondert hinterlegt, um ein vollständiges Parsen des Payloads zu vermeiden.
Der Nachrichten-Typ wird über das physische Ordner-Layout kodiert und benötigt keine zusätzliche redundante Abbildung.

```
- times/
    + 2026/ # year which contains one file per day binary files
        + 04_02.bin  # one file per day, e.g. 02.April
                     # monotonic and sorted: append-only
                     # []timestamp|sequence id 
  
- events/   
    + 1/
      - schema.json
      - acl.json
      - 0_999123.bin # min-inc to max-inc sequence id
      - 999123_.bin # pending segment file
      - ...
    

```

Die Performance-Abwägungen sehen wie folgt aus:

* Die Entwickler und Domänenexperten können im Zuge des Entdeckens und schnellen Iterierens von Implementierungen in der Domäne grundsätzlich selten glaubhafte Topic oder Partitionierungskriterien abschätzen. Nur die Eigenschaften Event-Typ, Sequenz-ID und Zeitstempel stehen als Zugriffs- oder Partitionsmerkmal verlässlich zur Verfügung, variieren aber mitunter in Menge und Dichte erheblich. 
* Es wird angenommen, dass Subscriber vor allem eine Replay-Mechanik mit einem Min- und Max-Offset der Sequenz-ID durchführen.
Etwaige persistente Cursor-Semantiken kann aus Konsistenzgründen nur der Subscriber selbst verwalten.
* Es wird ebenfalls angenommen, dass häufig nur eine kleine Teilmenge aller möglichen Event-Typen in einem Replay relevant ist.
Durch die Fan-Out Verzeichnisse können die entsprechenden Events ohne jegliches IO direkt ausgelassen werden und weder die Menge an hinterlegten Event-Typen noch die gespeicherte Menge an Events von ignorierten Events verursacht irgendeine Performanceeinbuße im Replay.
* Jedes Event-Typ Verzeichnis kann eine beliebige Menge an Dateien enthalten, die die eigentlichen Events kodieren.
Der Dateiname enthält die minimal und die maximal Sequenz-Nummer.
Wann eine neue Datei begonnen wird, obliegt der jeweiligen Optimierungsstrategie.
So macht es Sinn, dass wenn eine Retention-Strategie nach Tagen gewählt wird, dass die enthaltene Sequenz-ID bereits auch nach Tagen gruppiert wurden und mit einer I/O Operation aus dem Dateisystem entfernt werden können.
* Alternativ kann ein Client sich zu einem Zeitstempel die kleinste Sequenz-ID geben lassen.
Es ist anzunehmen, dass dies ein eher seltener oder explorativer Fall ist und kein reguläres Query-Verhalten darstellt.
Daher ist ein grob vorgeclusterter Inverser-Index vorgesehen, der die nicht eindeutigen aber monoton steigenden Zeitstempel nach Tagen sortiert abspeichert, um diese immer noch in O(log(n)) ohne In-Memory Datenstruktur aufzulösen.
Das Hinzufügen ist damit in O(1) lösbar.
* Ein beliebiges Löschen ist grundsätzlich nicht vorgesehen.
Stattdessen werden die Typen der zu löschenden Events entsprechend auf gelöscht gesetzt und der Payload mit Nullen überschrieben.
Eine optionale Compaction-Phase könnte dann die komplette Sequenz-Datei neu schreiben und dadurch den nicht mehr benötigten Speicher freigeben.
Die monotone Ordnung der Sequenznummern muss jedoch immer beibehalten werden und freigegebener Speicher darf nicht mit neueren Events aufgefüllt werden.
Frei gewordenen Sequenz-ID dürfen niemals erneut ausgeteilt werden.
* Die nächste ID kann durch ein einfaches Listing aller Event-Typ-Verzeichnisse und den darin enthaltenen Segment-Dateien zum Start der Datenbank erfolgen.
* Das Speichern eines neuen Events ist immer in O(1) durch das Anhängen an eine Segment-Datei im jeweiligen Event-Typ-Ordner möglich.
Wird ein Split-Kriterium angewendet, wird das Event in eine neue pending-Segmentdatei geschrieben.

## Konsistenz

Die folgenden Annahmen zur Konsistenz werden getroffen

* Der Event-Store insgesamt darf nur von einem Prozess bzw. einer Storeinstanz zur Zeit geöffnet sein.
Dies wird mittels flock Implementierung zur Laufzeit sichergestellt.
* Auf fsync wird grundsätzlich verzichtet, da der Service in einer virtuellen Umgebung betrieben wird, bei der weder Hard- noch Software bekannt sind.
Theoretische wie praktische Untersuchen zeigen, dass fsync kein Garant von Konsistenz ist. 
Andererseits zeigt der Betrieb von Services in der Praxis keine Auffälligkeiten diesbezüglich (Hetzner Cloud-Server).
Es wird preadAt und pwriteAt verwendet, um Events in einem Zug in einem konsistenten Buffer zu lesen und zu schreiben.
Darin enthalten ist auch eine Prüfsumme des gesamten Eintrags.
Somit können einzelne korrupte Einträge oder Abrisse am Dateiende erkannt und ignoriert werden.
Bei besonders starken Datei-Korruptions tritt im Zweifel ein Verlust der kompletten Segment-Datei ein, jedoch nie ein vollständiger Verlust, da alle nicht betroffenen Segment-Dateien intakt sind und unabhängig gelesen werden können.
Die Menge an Datenverlust kann also auch durch das Split-Kriterium bestimmt werden.
Es wird eine optionale Event-Erweiterung vorgesehen, mit der ein fsync explizit für bestimmte Events herbeigeführt werden kann, um besonders schützenswerte Daten direkt zu committen.
* Das Produktiv-System ist auf Ubuntu LTS Versionen auf einem virtuellen Server beschränkt.
Ein Produktiv-Betrieb auf MacOS oder Window oder anderen Linux Versionen ist ausgeschlossen.
* Die Compaction von Segmentdateien oder das Erstellen einer Segmentdatei auf Basis einer Pending-Datei als auch das Schreiben von Konfigurationsdateien erfordert immer die Erstellung innerhalb einer temporären Datei und ein atomic-Rename. 
* Die fehlende Schema-Validierung wird akzeptiert und durch eine aufgesetzte Schema-Verwaltung mit Berechtigungen und ACL-Token-Verwaltung sichergestellt, die in der jeweiligen Organisation per Anweisung vorgegeben und eingehalten werden sollte. 
* Data-Korruption-Fälle sollen als Fehler zurückgegeben oder geloggt werden, aber den Weiterbetrieb im Allgemeinen nicht blockieren.
Der konsistente Weiterbetrieb wird höher priorisiert, als die Wiedergabe aller Events, damit beispielsweise Bitrot-Fehler nicht zu einem Stillstand der Datenverarbeitung führen.

## Skalierbarkeit

Das System skaliert nur vertikal und benötigt schnellen SSD Speicher mit kurzen Zugriffszeiten.
Langsame File-Index Operationen werden durch den Prozess gechached und nur bei Programmstart nach und nach aufgebaut, damit der Prozessstart keine lineare Index-Phase erfordert.
Das System unterstützt ein File-Pooling von offenen Dateien, um etwaige Segment-Mengen im Millionenbereich (wahrscheinlich eher ein Konfigurationsfehler) zu verwalten.
Als Speicher werden konventionelle SSD RAIDs oder SSD-ZFS empfohlen.
Cluster-Dateisysteme wie CEPH sind ebenfalls denkbar, eröffnen aber wieder diverse Konsistenzbedenken.

Durch den global streng monotonen Zähler und die Beschränkung auf einen Prozess, ergibt sich hier das obere Limit an Belastung.
Durch die geschickte Wahl der Segmentierungs- und Retention Bedingung ist allerdings anzunehmen, dass das System eine kontinuierliche konstante Schreiblast bewältigen kann, die nur durch absolute Maschinenleistung limitiert wird, aber nicht durch die Datenmenge als solche.
Eine stabile Einfügerate von 500tsd Kleinst-Events (wenige Bytes pro Event) pro Sekunde erscheint somit grundsätzlich erreichbar.


## Benchmarks

Die folgenden Ergebnisse wurden auf einem Apple M1 Max (macOS, arm64, 10 Cores) gemessen.
Die Benchmarks können mit `go test -bench=. -benchmem ./pkg/ndb/msgstore/` reproduziert werden.

### Schreiben (Append, ohne Kompression)

```
| Payload   | msg/s    | MB/s      | ns/op  | allocs/op |
+-----------+----------+-----------+--------+-----------+
| 0 B       | 351.685  |       –   |  2.843 |         5 |
| 64 B      | 349.738  |      22,4 |  2.859 |         5 |
| 256 B     | 328.013  |      84,0 |  3.049 |         5 |
| 1 KiB     | 260.704  |     267,0 |  3.836 |         5 |
| 4 KiB     | 144.455  |     591,7 |  6.923 |         5 |
| 16 KiB    |  88.290  |   1.446,5 | 11.326 |         5 |
```

### Schreiben mit S2-Kompression

```
| Payload   | msg/s    | MB/s      | allocs/op |
+-----------+----------+-----------+-----------+
| 256 B     | 138.538  |      35,5 |         6 |
| 1 KiB     | 129.074  |     132,2 |         6 |
| 4 KiB     | 127.354  |     521,6 |         6 |
| 16 KiB    | 105.263  |   1.724,6 |         6 |
```

### Konkurrentes Schreiben (64 B Payload, N Writer auf N Event-Typen)

Jeder Writer schreibt ausschließlich in seinen eigenen Event-Typ (separate Segment-Datei).
Die Sequenz-ID wird lock-free per `atomic.AddUint64` vergeben, der per-Type-Mutex serialisiert
nur Appends auf denselben Typ. Der Time-Index wird in einem 1 MiB In-Memory-Buffer gepuffert
und erst bei Schwellwert, Tageswechsel oder Lookup geflusht – dadurch entfällt der pwrite-Syscall
pro Nachricht als Flaschenhals und die Schreibrate skaliert mit der Anzahl der Writer.

```
| Writer | msg/s (total) | ns/op  | Speedup vs. 1 Writer                  |
+--------+---------------+--------+----------------------------------------+
|      1 |       336.000 |  2.974 | Baseline                               |
|      2 |       472.000 |  2.090 | 1,4×                                   |
|      4 |       544.000 |  1.837 | 1,6×                                   |
|      8 |       529.000 |  1.890 | 1,6× – FilePool-Mutex wird Bottleneck  |
```

### Lesen (Replay, ohne Kompression)

```
| Payload   | msg/s     | MB/s      | allocs/op |
+-----------+-----------+-----------+-----------+
| 0 B       |   722.978 |       –   |        33 |
| 64 B      |   720.246 |      46,1 |        33 |
| 256 B     |   701.143 |     179,5 |        33 |
| 1 KiB     |   625.426 |     640,4 |        33 |
| 4 KiB     |   426.401 |   1.746,5 |        68 |
```

### k-Way-Merge Replay über mehrere Event-Typen

```
| Typen | msg/s     | allocs/op |
+-------+-----------+-----------+
|     1 |   714.634 |        39 |
|     5 |   671.125 |       158 |
|    20 |   612.923 |       583 |
```

### Serialisierung (CPU-Obergrenze, ohne I/O)

```
| Operation   | 0 B msg/s | 64 B msg/s | 1 KiB msg/s | 4 KiB msg/s |
+-------------+-----------+------------+-------------+-------------+
| Marshal     |    73 Mio |     59 Mio |     5,3 Mio |     1,4 Mio |
| Unmarshal   |    69 Mio |     51 Mio |     6,9 Mio |     1,6 Mio |
```

Die Schreibleistung wird primär durch den pwriteAt-Syscall pro Event-Append limitiert.
Der Time-Index wird in einem 1 MiB Buffer gepuffert und nur bei Bedarf geflusht, sodass er
kein Bottleneck mehr darstellt.
Die Leseleistung skaliert nahezu linear mit der Payload-Größe und ist durch die zero-copy Deserialisierung und den File-Pool weitgehend allokationsfrei.

## File specification 

### Segment File Format

Filename specification

```
<minSeq>_<maxSeq>.bin     # finalized segment
<minSeq>_.bin             # pending/open segment
```

Segment file header specification

```
| Field          | Bytes  | Description                                |
+----------------+--------+--------------------------------------------+
| Magic          | 4      | 0x4E41474F ("NAGO")                        |
| Format Magic   | 4      | 0x4E455653 ("NEVS") Nago Event Store       |
| Version        | 1      | Format version, z.B. 0x01                  |
| Messages       | n      | Append only messages                       |
+----------------+--------+--------------------------------------------+
```

### Message Format in Segment File v1

Each message is wrapped in a frame that starts with a fixed 8-byte sync marker followed by the inner message length.
The sync marker enables forward-scanning recovery after corruption (bitrot, partial writes).

Note that the message type identifier is derived from the parents folder name and is not encoded redundantly.

```
| Field          | Bytes  | Description                                |
+----------------+--------+--------------------------------------------+
| SyncMarker     | 8      | 0xDEAD4E455653BEEF – resync anchor         |
| TotalLen       | 4      | byte length of the inner message below     |
| SequenceID     | 8      | global strict monotonic, 0=tombstone       |
| Timestamp      | 8      | append unix timestamp in nano seconds      |
| TraceID        | 16     | random id, like a UUID                     |
| Encoding       | 1      | 0=raw, 1=S2 Compression                    |
| PayloadLen     | 4      | byte length of the payload                 |
| UnompressedLen | 4      | uncompressed byte length of the payload    |
| Payload        | n      | raw payload bytes                          |
| CRC32          | 4      | CRC of SequenceID..Payload (inner message) |
+----------------+--------+--------------------------------------------+
```

TotalLen equals msgFixedSize (45) + PayloadLen and is cross-validated during deserialization.

#### Recovery semantics

* **Bitrot / mid-file corruption**: When a CRC mismatch, invalid sync marker, or implausible TotalLen/PayloadLen is detected, the reader scans forward byte-by-byte for the next valid sync marker and continues from there. Corrupt messages are logged and skipped.
* **Tail truncation (crash recovery)**: When the file ends with an incomplete message (partial write due to crash), the trailing bytes are silently ignored during reads. For pending segments, the incomplete tail is physically truncated on next open so that subsequent appends start at a clean boundary.

To protect against wrong usage or attacks, the maximum allowed message size can be configured, but defaults to 16MiB.
Also, the transparent compression can be enabled and is extensible for future algorithms.
