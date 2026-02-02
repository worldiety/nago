#!/bin/bash
# Build-Skript für Nago Security Audit Documentation
# Copyright (c) 2025 worldiety GmbH

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/output"

echo "=== Nago Security Documentation Build ==="
echo ""

# Output-Verzeichnis erstellen
mkdir -p "$OUTPUT_DIR"

# Prüfen ob asciidoctor-pdf installiert ist
if ! command -v asciidoctor-pdf &> /dev/null; then
    echo "FEHLER: asciidoctor-pdf ist nicht installiert."
    echo "Installieren Sie es mit: gem install asciidoctor-pdf"
    exit 1
fi

# Prüfen ob PlantUML verfügbar ist
PLANTUML_AVAILABLE=false
if command -v plantuml &> /dev/null; then
    PLANTUML_AVAILABLE=true
    echo "✓ PlantUML gefunden"
else
    echo "⚠ PlantUML nicht gefunden - Diagramme werden nicht gerendert"
    echo "  Installieren Sie es mit: brew install plantuml (macOS)"
fi

# Prüfen ob asciidoctor-diagram verfügbar ist
DIAGRAM_AVAILABLE=false
if gem list asciidoctor-diagram -i &> /dev/null; then
    DIAGRAM_AVAILABLE=true
    echo "✓ asciidoctor-diagram gefunden"
else
    echo "⚠ asciidoctor-diagram nicht gefunden"
    echo "  Installieren Sie es mit: gem install asciidoctor-diagram"
fi

echo ""
echo "=== Generiere PDF ==="

# PDF generieren
if [ "$DIAGRAM_AVAILABLE" = true ] && [ "$PLANTUML_AVAILABLE" = true ]; then
    echo "Generiere PDF mit eingebetteten Diagrammen..."
    asciidoctor-pdf \
        -r asciidoctor-diagram \
        -a pdf-theme=default \
        -a pdf-fontsdir=GEM_FONTS_DIR \
        -a allow-uri-read \
        -a imagesoutdir="$OUTPUT_DIR/images" \
        -D "$OUTPUT_DIR" \
        -o nago-security-documentation.pdf \
        "$SCRIPT_DIR/security-overview.adoc"
else
    echo "Generiere PDF ohne Diagramme..."
    asciidoctor-pdf \
        -a pdf-theme=default \
        -D "$OUTPUT_DIR" \
        -o nago-security-documentation.pdf \
        "$SCRIPT_DIR/security-overview.adoc"
fi

echo ""
echo "=== Generiere separate Kapitel-PDFs ==="

# Einzelne Kapitel als separate PDFs
for file in "$SCRIPT_DIR"/security-*.adoc; do
    filename=$(basename "$file" .adoc)
    if [ "$filename" != "security-overview" ]; then
        echo "  Generiere $filename.pdf..."
        if [ "$DIAGRAM_AVAILABLE" = true ] && [ "$PLANTUML_AVAILABLE" = true ]; then
            asciidoctor-pdf \
                -r asciidoctor-diagram \
                -a allow-uri-read \
                -a imagesoutdir="$OUTPUT_DIR/images" \
                -D "$OUTPUT_DIR" \
                -o "$filename.pdf" \
                "$file" 2>/dev/null || echo "    ⚠ Warnung bei $filename"
        else
            asciidoctor-pdf \
                -D "$OUTPUT_DIR" \
                -o "$filename.pdf" \
                "$file" 2>/dev/null || echo "    ⚠ Warnung bei $filename"
        fi
    fi
done

echo ""
echo "=== PlantUML Diagramme (SVG) ==="

if [ "$PLANTUML_AVAILABLE" = true ]; then
    mkdir -p "$OUTPUT_DIR/diagrams"

    # PlantUML aus AsciiDoc extrahieren und rendern
    for file in "$SCRIPT_DIR"/security-*.adoc; do
        if grep -q "\[plantuml" "$file"; then
            filename=$(basename "$file" .adoc)
            echo "  Extrahiere Diagramme aus $filename..."
            # PlantUML kann direkt aus AsciiDoc-Dateien extrahieren
            plantuml -tsvg -o "$OUTPUT_DIR/diagrams" "$file" 2>/dev/null || true
        fi
    done
else
    echo "  Übersprungen (PlantUML nicht verfügbar)"
fi

echo ""
echo "=== Build abgeschlossen ==="
echo ""
echo "Ausgabe-Verzeichnis: $OUTPUT_DIR"
echo ""
echo "Generierte Dateien:"
ls -la "$OUTPUT_DIR"/*.pdf 2>/dev/null || echo "  Keine PDFs gefunden"
echo ""

if [ -d "$OUTPUT_DIR/diagrams" ]; then
    echo "Diagramme:"
    ls -la "$OUTPUT_DIR/diagrams"/*.svg 2>/dev/null || echo "  Keine Diagramme gefunden"
fi

echo ""
echo "Fertig!"
