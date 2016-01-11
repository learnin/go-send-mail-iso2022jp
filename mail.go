package mail

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type SmtpClient struct {
	Host     string
	Port     uint16
	Username string
	Password string
	client   *smtp.Client
}

type Mail struct {
	From    string
	To      string
	Subject string
	Body    string
}

func (c *SmtpClient) Connect() error {
	client, err := smtp.Dial(fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return err
	}
	if err := client.Hello("localhost"); err != nil {
		return err
	}
	if c.Username != "" {
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(smtp.PlainAuth("", c.Username, c.Password, c.Host)); err != nil {
				return err
			}
		}
	}
	c.client = client
	return nil
}

func (c *SmtpClient) Close() error {
	return c.client.Close()
}

func (c *SmtpClient) Quit() error {
	return c.client.Quit()
}

func (c *SmtpClient) SendMail(m Mail) error {
	if err := c.client.Reset(); err != nil {
		return err
	}
	isNameAddrFrom := false
	var envelopeFrom string
	var fromName string
	r, err := regexp.Compile("^(.*)<(.*)>$")
	if err != nil {
		return err
	}
	if match := r.FindStringSubmatch(m.From); match != nil {
		isNameAddrFrom = true
		fromName = match[1]
		envelopeFrom = match[2]
	} else {
		envelopeFrom = m.From
	}
	c.client.Mail(envelopeFrom)
	if err := c.client.Rcpt(m.To); err != nil {
		return err
	}
	w, err := c.client.Data()
	if err != nil {
		return err
	}
	var headerFrom string
	if isNameAddrFrom {
		name, err := encodeHeader(fromName)
		if err != nil {
			return err
		}
		headerFrom = name + " <" + envelopeFrom + ">"
	} else {
		headerFrom = envelopeFrom
	}

	subject, err := encodeHeader(m.Subject)
	if err != nil {
		return err
	}
	body, err := encodeToJIS(m.Body + "\r\n")
	if err != nil {
		return err
	}
	msg := "From: " + headerFrom + "\r\n" +
		"To: " + m.To + "\r\n" +
		"Subject:" + subject +
		"Date: " + time.Now().Format(time.RFC1123Z) + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=ISO-2022-JP\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n" +
		"\r\n" +
		body
	if _, err = w.Write([]byte(msg)); err != nil {
		return err
	}
	return w.Close()
}

func encodeToJIS(s string) (string, error) {
	r, err := ioutil.ReadAll(transform.NewReader(strings.NewReader(s), japanese.ISO2022JP.NewEncoder()))
	if err != nil {
		return "", err
	}
	return string(r), nil
}

func encodeHeader(subject string) (string, error) {
	b := make([]byte, 0, utf8.RuneCountInString(subject))
	for _, s := range splitByCharLength(subject, 13) {
		b = append(b, " =?ISO-2022-JP?B?"...)
		s, err := encodeToJIS(s)
		if err != nil {
			return "", err
		}
		b = append(b, base64.StdEncoding.EncodeToString([]byte(s))...)
		b = append(b, "?=\r\n"...)
	}
	return string(b), nil
}

func splitByCharLength(s string, length int) []string {
	result := []string{}
	b := make([]byte, 0, length)
	for i, c := range strings.Split(s, "") {
		b = append(b, c...)
		if i%length == 0 {
			result = append(result, string(b))
			b = make([]byte, 0, length)
		}
	}
	if len(b) > 0 {
		result = append(result, string(b))
	}
	return result
}
