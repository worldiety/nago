package magic

import "bytes"

var zipMagicBytes = [][]byte{
	{0x50, 0x4B, 0x03, 0x04}, // Normale ZIP-Datei
	{0x50, 0x4B, 0x05, 0x06}, // Leere ZIP-Datei (End of Central Directory)
	{0x50, 0x4B, 0x07, 0x08}, // Spanned ZIP-Datei
}

// Detect returns the estimated mimetype of the given buffer.
func Detect(buf []byte) string {
	if bytes.HasPrefix(buf, []byte("%PDF-")) {
		return "application/pdf"
	}

	for _, m := range zipMagicBytes {
		if bytes.HasPrefix(buf, m) {
			return "application/zip"
		}
	}

	return "application/octet-stream"
}

// Ext returns the typical estimated filename extensions
func Ext(buf []byte) string {
	switch Detect(buf) {
	case "application/pdf":
		return ".pdf"
	case "application/zip":
		return ".zip"
	default:
		return ".bin"
	}
}
