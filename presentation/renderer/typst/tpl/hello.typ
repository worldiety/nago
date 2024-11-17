#set text(
  weight: "extralight",
  font: (
    "Helvetica Neue"
  ),
)

#set text(lang: "de")

#set page(
  paper: "a4",
  header: align(left)[
    #pad(top: 2em, image("wdy.png"))
  ],
  margin: (
    top: 3cm,
    bottom: 4cm,
    x: 2.5cm,
  ),
)

#show heading.where(level: 1): it => [
  #set align(left)
  #set text(18pt, weight: "extralight", fill: rgb("#1B8C30"))
  #block(above: 22pt, below: 22pt, smallcaps(it.body))
]

#show heading.where(level: 2): it => [
  #set align(left)
  #set text(18pt, weight: "extralight", fill: rgb("#1B8C30"))
  #block(above: 22pt, below: 22pt, smallcaps(it.body))
]

#show heading.where(level: 3): it => [
  #set align(left)
  #set text(18pt, weight: "extralight", fill: rgb("#1B8C30"))
  #block(above: 22pt, below: 22pt, smallcaps(it.body))
]

// Medium bold table header.
#show table.cell.where(y: 0): it => {
  // not yet implemented, see https://github.com/typst/typst/issues/4159
  set table.cell(fill: rgb("#1B8C30"))
  set text(weight: "regular", fill: rgb("#ffffff"))
  upper(it)
}

// workaround, see https://github.com/typst/typst/issues/4159
#set table(fill: (x, y) => if y == 0 { rgb("#1B8C30") })
#set table(stroke: 0.5pt + rgb("#1B8C30"))
#set table(stroke: (x, y) => (
  left: if x > 0 { 0.5pt + rgb("#1B8C30")},
  bottom: if y > 0 { 0.5pt + rgb("#1B8C30")},
))

#show outline.entry.where(
  level: 2,
): it => [ #text(style: "italic", weight: "bold", it)]

#align(center, block[
  #image("teaser_app.png", height: 70%)

  #box(
    height: 2em,
  )

  #box(
    width: 15cm,
    [
      #text(
        size: 20pt,
        fill: rgb("#1B8C30"),
        [Veranstaltungs-App],
      )

      Erstellung eines Minium Viable Products (MVP) & Ausbaustufen
    ],
  )
])

#pagebreak()

#counter(page).update(1)

#set page(
  paper: "a4",
  header: [
    #grid(
      columns: (2fr, 1fr),
      rows: 1,
      grid.cell()[ #pad(top: 2em, image("wdy.png"))],
      grid.cell(align: right)[
        #set text(size: 8pt, fill: rgb("#7F7F7F"))
        #grid(
          columns: (1fr,1fr),
          gutter: 3pt,
          grid.cell(align: left)[Seite],
          grid.cell(align: right)[ #context { counter(page).display("1/1",both: true)} ],
          grid.cell(align: left)[Datum],
          grid.cell(align: right)[16.11.2024],
          grid.cell(align: left)[Dokument],
          grid.cell(align: right)[2024111601],
        )
      ],
    )
  ],
  footer: [
    #set text(fill: rgb("#7F7F7F"))

    worldiety GmbH
    #place(dx: -3cm, dy: -0.2cm)[ #line(length: 28cm, stroke: rgb("#1B8C30"))]
    #set text(size: 8pt)

    #grid(
      columns: (0.7fr,1fr,1fr,1fr),
      gutter: 3pt,
      grid.cell(align: left)[
        Nordseestraße 2 \
        26131 Oldenburg\
        Deutschland
      ],
      grid.cell(align: left)[
        Registergericht Oldenburg (Oldb.) \
        HRB 208428 / Geschäftsführer \
        Adrian Macha & Torben Schinke \
      ],
      grid.cell(align: left)[
        www.worldiety.com \
        info\@worldiety.com \
        +49 441 559 770 0 \
      ],
      grid.cell()[],
    )
  ],
)

#text(size: 9pt)[
  #underline[worldiety GmbH, Nordseestraße 2, 26131 Oldenburg]\
]

#text(size: 9pt)[
  Handwerkskammer Oldenburg\
  Herr Kai Vensler\
  Theaterwall 32\
  26122 Oldenburg\
]

= Entwicklung einer Messe-App und Verwaltungsplattform

Sehr geehrter Herr Vensler,

gerne bieten wir Ihnen an, mit Ihnen eine Begleit-App für den neu aufgesetzten „Talenttag Handwerk“ zu entwickeln, welche von Schülern und Schülerinnen, sowie Lehrenden ausgewähl-ter Schulen zur Planung des Veranstaltungstages genutzt werden kann.

Die zugrundeliegenden Daten können Sie über eine Webanwendung komfortabel verwalten.

Auf den folgenden Seiten finden Sie eine Auflistung unserer Leistungen.

Wir freuen uns auf eine erfolgreiche Zusammenarbeit!

Mit freundlichen Grüßen

#image("sig.png", height: 1cm)
#place(dy: -0.8cm)[#line(length: 3cm)]
#text(weight: "bold")[Julia Schloen]\

Projektmanagerin\
+49 441 559 770 12\
#text(fill: rgb("#1B8C30"))[julia.schloen\@worldiety.de]\

#pagebreak()

#outline(indent: 1cm)

#pagebreak()

= Kurzvorstellung worldiety

Als Oldenburger- IT-Dienstleister hat worldiety sich auf die Beratung sowie Entwicklung von individuellen
Softwarelösungen spezialisiert. Seit mehr als 10 Jahren unterstützen wir unsere Kunden auf ihrem Weg in
die digitale Transformation und sorgen dafür, dass sich durch unsere Software-lösungen die Steuerungs- sowie
Prozesseffizienz in Ihrem Unternehmen erhöht.

Hierfür entwickeln wir individuelle Softwarelösungen und stehen Ihnen in unterschiedlichen Be-reichen entlang des Softwarelebenszyklus mit unserer langjährigen Expertise jederzeit als Partner zur Seite:

- Kickstarter – ein erster unverbindlicher Workshop zur Erkennung Ihres individuellen Digita-lisierungspotentials
- IT-Beratung – für Ihre Digitalisierungsstrategien, moderne Geschäftsprozesse und einen perfekten Start in die digitale Transformation
- Softwareentwicklung – auf Ihre Bedürfnisse individuell entwickelte Softwarelösungen in den Bereichen Mobile, Web, Cloud, E-Commerce und Unternehmenssoftware
- Qualitätsmanagement – für qualitativ hochwertige Ergebnisse Ihrer digitalen Prozesse sowie Softwareprodukte
- Betrieb & Wartung – die Sicherstellung der Wertschöpfungskette Ihres Produkts durch re-gelmäßige Weiterentwicklungen, Wartungsarbeiten & Hosting

Innerhalb unserer fünf Leistungsbereiche beschäftigen wir uns mit innovativen Themen wie Cross-Plattform-Entwicklung, Headless Content Management Systemen,
Machine Learning oder der Orchestrierung von skalierbaren (Microservice-) Architekturen.

Als qualifizierter Partner stehen wir im Bereich Data Analytics unterstützend zur Seite und können so gemeinsam neue Marktpo-tenziale erschließen.
Wir analysieren die individuellen Bedürfnisse unserer Kunden und entwerfen oder lizenzieren geeignete digitale Lösungen. Da mit der Entwicklung einer Software die Arbeit noch nicht abge-schlossen ist, sichert unser zertifiziertes Qualitätsmanagement die Qualität der Software und wird im Anschluss von unserem Betrieb & Wartungs-Service kontinuierlich gepflegt.

Als ein autorisiertes Beratungsunternehmen für verschiedene Förderprogramme im Bereich der Digitalisierung stehen wir mit unserer Digitalkompetenz jederzeit beratend zur Seite.

worldiety – #text(style: "italic")[for a digital world and mobile society]

== third level

Blub

#table(
  columns: (0.4fr, 1fr, 1fr, 1fr),
  table.header[Month][Title][Author][Genre],
  [January],
  [The Great Gatsby],
  [F. Scott Fitzgerald],
  [Classic],
  [February],
  [To Kill a Mockingbird],
  [Harper Lee],
  [Drama],
  [March],
  [1984],
  [George Orwell],
  [Dystopian],
  [April],
  [The Catcher in the Rye],
  [J.D. Salinger],
  [Coming-of-Age],
  table.cell(colspan: 2, rowspan: 2)[jo man],
  [A],
  [B],
  [C],
  [D],
)
