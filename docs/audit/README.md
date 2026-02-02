# Nago Security Audit Documentation

Diese Dokumentation beschreibt die Sicherheitsarchitektur, implementierte Kontrollen und Compliance-Nachweise des Nago Low-Code Frameworks für Sicherheitsaudits nach ISO 27001 und Penetrationstests.

## Dokumentstruktur

| Datei | Beschreibung |
|-------|--------------|
| `security-overview.adoc` | Hauptdokument mit Executive Summary und Glossar |
| `security-architecture.adoc` | Systemarchitektur, Authentifizierung, Kryptographie |
| `security-controls.adoc` | ISO 27001 Annex A Controls Mapping |
| `security-threats.adoc` | STRIDE Bedrohungsmodellierung |
| `security-pentest-checklist.adoc` | OWASP Top 10 Pentest-Checkliste |
| `security-compliance.adoc` | Compliance-Nachweise (ISO 27001, SOC 2, DSGVO) |
| `security-incident-response.adoc` | Incident Response Plan |

## Quick Start

### Für ISO 27001 Auditoren

1. Beginnen Sie mit `security-overview.adoc` für die Executive Summary
2. Prüfen Sie `security-controls.adoc` für das Statement of Applicability
3. Nutzen Sie `security-compliance.adoc` für detaillierte Evidenzen

### Für Penetrationstester

1. Lesen Sie `security-architecture.adoc` für die Systemübersicht
2. Nutzen Sie `security-pentest-checklist.adoc` für Testszenarien
3. Beachten Sie `security-threats.adoc` für bekannte Risiken

## PDF-Generierung

### Voraussetzungen

```bash
# asciidoctor-pdf installieren
gem install asciidoctor-pdf

# PlantUML installieren (für Diagramme)
brew install plantuml  # macOS
# oder
apt-get install plantuml  # Ubuntu/Debian
```

### PDF erstellen

```bash
# Einzelnes PDF mit allen Kapiteln
asciidoctor-pdf security-overview.adoc -o nago-security-documentation.pdf

# Mit PlantUML-Diagrammen
asciidoctor-pdf -r asciidoctor-diagram security-overview.adoc -o nago-security-documentation.pdf
```

### PlantUML-Diagramme rendern (optional)

```bash
# Falls Diagramme separat gerendert werden sollen
plantuml -tsvg security-architecture.adoc
plantuml -tsvg security-threats.adoc
plantuml -tsvg security-incident-response.adoc
```

## Build-Skript

Ein vollständiges Build-Skript finden Sie in `build.sh`:

```bash
chmod +x build.sh
./build.sh
```

## Versionierung

| Version | Datum | Änderungen |
|---------|-------|------------|
| 1.0 | 2026-02-02 | Initiale Version für ISO 27001 Audit |

## Kontakt

| Anfragen | Kontakt |
|----------|---------|
| Security Issues | security@worldiety.de |
| Compliance-Fragen | compliance@worldiety.de |
| Technische Dokumentation | https://nago.dev |

## Deployment-Security

> **Hinweis:** Diese Dokumentation behandelt ausschließlich die **Application Layer Security** des Nago Frameworks.
>
> Für Deployment-bezogene Sicherheitsthemen (Infrastruktur, TLS, Firewall, Container) siehe die separate **Nago Hub Hosting Dokumentation**.

## Lizenz

Diese Dokumentation ist Teil des NAGO Low-Code Frameworks und unterliegt der Nago-Lizenz.

Copyright (c) 2026 worldiety GmbH
