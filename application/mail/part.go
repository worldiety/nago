package mail

import (
	"bytes"
	"fmt"
	"io"
	"net/mail"
	"strings"
)

type Part struct {
	Header  mail.Header
	Encoded []byte
}

func (p *Part) PutHeader(key, value string) {
	if p.Header == nil {
		p.Header = make(mail.Header)
	}

	p.Header[key] = append(p.Header[key], value)
}

func (p *Part) write(writer io.Writer) error {
	if p.Header != nil {
		for k, v := range p.Header {
			str := fmt.Sprintf("%s: %s\r\n", k, strings.Join(v, ";"))
			_, err := writer.Write([]byte(str))
			if err != nil {
				return err
			}
		}
		_, err := writer.Write([]byte("\r\n"))
		if err != nil {
			return err
		}
	}
	if p.Encoded != nil {
		_, err := writer.Write(p.Encoded)
		return err
	}
	return nil
}

func NewTextPart(msg string) Part {
	part := Part{}
	part.PutHeader("Content-Type", "text/plain")
	part.PutHeader("Content-Type", "charset=utf-8")
	part.PutHeader("Content-Type", "format=flowed")
	part.PutHeader("Content-Transfer-Encoding", "8bit")

	var tmp bytes.Buffer
	if err := writeText(msg, &tmp); err != nil {
		panic(fmt.Errorf("unreachable: %w", err))
	}

	part.Encoded = tmp.Bytes()

	return part
}

func NewHtmlPart(html string) Part {
	part := Part{}
	part.PutHeader("Content-Type", "text/html")
	part.PutHeader("Content-Type", "charset=utf-8")
	part.PutHeader("Content-Transfer-Encoding", "8bit")
	var tmp bytes.Buffer
	if err := writeText(html, &tmp); err != nil {
		panic(fmt.Errorf("unreachable: %w", err))
	}

	part.Encoded = tmp.Bytes()
	return part
}

func NewAttachmentPart(name string, data []byte) Part {
	part := Part{}
	part.PutHeader("Content-Type", "application/octet-stream")
	part.PutHeader("Content-Type", "name="+protect(name))
	part.PutHeader("Content-Transfer-Encoding", "base64")

	var tmp bytes.Buffer
	for _, line := range b64(data) {
		_, err := tmp.Write([]byte(line))
		if err != nil {
			panic(fmt.Errorf("unreachable: %w", err))
		}
		_, err = tmp.Write([]byte("\r\n"))
		if err != nil {
			panic(fmt.Errorf("unreachable: %w", err))
		}

	}

	part.Encoded = tmp.Bytes()
	return part
}
