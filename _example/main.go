package main

import (
	"github.com/learnin/go-send-mail-iso2022jp"
)

func main() {
	c := mail.SmtpClient{
		Host:     "localhost",
		Port:     1025,
		Username: "",
		Password: "",
	}
	if err := c.Connect(); err != nil {
		panic(err)
	}
	defer func() {
		c.Close()
		c.Quit()
	}()
	c.SendMail(mail.Mail{
		From:    "ほげ <sender@example.org>",
		To:      "ふー <receipt@example.org>",
		Subject: "テストおおおおおお。あああいいいいんんんaいいう1234あああああああああああいいいいいいいいいいう",
		Body:    "テスト本文",
	})
}
