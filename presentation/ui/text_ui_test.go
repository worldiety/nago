package ui

import "fmt"

// Example-Img: [/images/components/basic/text/link-example.png].
func ExampleLinkWithAction() {
	LinkWithAction("Nago Docs", func() {
		fmt.Printf("Nago is easy to use")
	})
	//Output:
}

// Example-Img: [/images/components/basic/text/link-example.png].
func ExampleLink() {
	Link(nil, "Nago Docs", "https://www.nago-docs.com", "_blank")
	//Output:
}

// Example-Img: [/images/components/basic/text/mail-to-example.png].
func ExampleMailTo() {
	MailTo(nil, "Worldiety", "info@worldiety.de")
	//Output:
}
