// Code generated by NAGO nprotoc DO NOT EDIT.

package {{.PackageName}}

import(
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"
)

{{.BinaryWriter}}

{{range .Marshals}}
{{.}}
{{end}}