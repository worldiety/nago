package mail

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"go.wdy.de/nago/pkg/data"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type SmtpRepository data.Repository[Smtp, SmtpID]

type SmtpID string

type Category string

const (
	Primary  Category = "primary"
	Normal   Category = "normal"
	Disabled Category = "disabled"
)

type Smtp struct {
	ID       SmtpID   `visible:"false"`
	Name     string   `label:"Name"`
	Host     string   `label:"Host" supportingText:"beispielsweise mail.your-server.de"`
	Port     int      `label:"Port" supportingText:"Der Standardport ist üblicherweise 587 mit TLS" table-visible:"false"`
	User     string   `label:"Login" table-visible:"false"`
	Password string   `label:"Passwort" style:"secret" table-visible:"false"`
	Category Category `label:"Kategorie" values:"[\"primary=Primär\", \"normal=Normal\",\"disabled=Deaktiviert\"]"`
}

func (s Smtp) WithIdentity(id SmtpID) Smtp {
	s.ID = id
	return s
}

func (s Smtp) Identity() SmtpID {
	return s.ID
}

func (s Smtp) send(m Mail) error {

	// Connect to the SMTP Server
	servername := s.Host + ":" + strconv.Itoa(s.Port)

	host, _, _ := net.SplitHostPort(servername)

	auth := smtp.PlainAuth("", s.User, s.Password, host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}

	conn, err := net.DialTimeout("tcp", servername, 10*time.Second)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	err = c.StartTLS(tlsconfig)
	if err != nil {
		return err
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		return err
	}
	if len(m.From.Address) == 0 {
		m.From.Address = s.User
	}

	// the from address is usually important for authentication
	if err = c.Mail(m.From.Address); err != nil {
		return err
	}

	// add all recipients, which is independent of what is in the actual message
	for _, adr := range m.To {
		if err = c.Rcpt(adr.Address); err != nil {
			return err
		}
	}

	for _, adr := range m.CC {
		if err = c.Rcpt(adr.Address); err != nil {
			return err
		}
	}

	for _, adr := range m.BCC {
		if err = c.Rcpt(adr.Address); err != nil {
			return err
		}
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	// Setup data
	boundaryMultipartMixed := "------------DD1AADA7899159F3F80A4C5A"
	data := &dataWriter{sb: &bytes.Buffer{}}
	data.writeHeader("From", m.From.String())
	data.writeHeader("To", recipients(m.To).String())
	data.writeHeader("CC", recipients(m.CC).String())
	data.writeHeader("Subject", mime.QEncoding.Encode("UTF-8", m.Subject))
	data.writeHeader("MIME-Version", "1.0")
	data.writeHeader("Content-Type", "multipart/mixed;  boundary=\""+boundaryMultipartMixed+"\"")
	data.rf()
	data.writeLine(" This is a multi-Part message in MIME format.")
	data.rf()
	data.rf()

	for _, p := range m.Parts {
		data.writeLine("--")
		data.writeLine(boundaryMultipartMixed)
		data.rf()
		err = p.write(data.sb)
		if err != nil {
			return err
		}

		data.rf()
	}

	data.writeLine("--")
	data.writeLine(boundaryMultipartMixed)
	data.writeLine("--")

	//fmt.Println(string(data.sb.Bytes()))
	_, err = w.Write(data.sb.Bytes())
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}

type recipients []mail.Address

func (r recipients) String() string {
	if len(r) == 0 {
		return ""
	}
	sb := &strings.Builder{}
	for i := 0; i < len(r)-1; i++ {
		sb.WriteString(r[i].String())
		sb.WriteString(",")
	}
	sb.WriteString(r[len(r)-1].String())
	return sb.String()
}

type dataWriter struct {
	sb *bytes.Buffer
}

func (d *dataWriter) writeHeader(key string, value string) *dataWriter {
	d.sb.WriteString(fmt.Sprintf("%s: %s\r\n", protect(key), protect(value)))
	return d
}

func (d *dataWriter) rf() *dataWriter {
	d.sb.WriteString("\r\n")
	return d
}

func (d *dataWriter) writeLine(str string) *dataWriter {
	d.sb.WriteString(str)
	return d
}
