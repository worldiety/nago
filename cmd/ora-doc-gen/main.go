package main

import (
	"fmt"
	"go.wdy.de/nago/presentation/protocol"
	"time"
)

func main() {
	generate()
}

func generate() {
	d := NewDoc()
	d.Printf("This documentation is auto-generated by ora-doc-gen. Do NOT edit.\n\n")
	d.Printf("Generated at %s.\n", time.Now().Format(time.DateTime))
	d.Printf(`

<style>
table th:first-of-type {
    width: 20%%;
}
table th:nth-of-type(2) {
    width: 30%%;
}
table th:nth-of-type(3) {
    width: 50%%;
}
table th:nth-of-type(4) {
    width: 30%%;
}
</style>

`)
	aboutChannel(d)
	d.Printf("## events\n\n")
	wTx(d)
	wAck(d)

	wConfigurationRequested(d)
	wConfigurationDefined(d)

	wNewComponentRequested(d)
	wComponentInvalidated(d)

	d.Printf("## Components\n\n")
	wButton(d)
	fmt.Println(d.out.String())
}

var ptr protocol.Ptr

func nextPtr() protocol.Ptr {
	ptr++
	return ptr
}

func nextRequestId() protocol.RequestId {
	ptr++
	return protocol.RequestId(ptr)
}

func nextCaption() string {
	return fmt.Sprintf("Caption No. %d", nextPtr())
}

func nextSVGSrc() protocol.SVGSrc {
	return protocol.SVGSrc(fmt.Sprintf("<svg>my inline svg %d</svg>", nextPtr()))
}

func nextUUID() string {
	return "3d159507-35a7-422b-9a77-74546bc5fcbe"
}
