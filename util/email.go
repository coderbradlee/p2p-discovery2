package util

import (
	"log"
	"strings"
	"time"

	"github.com/go-gomail/gomail"
)

var mailLastTime time.Time //邮件频率 上次发送时间

func AlertMail(model string, info ...string) {
	m := gomail.NewMessage()
	m.SetHeader("From", "hm@wident.cn")
	m.SetHeader("To", "huowutao@wident.cn", "leiweishi@wident.cn")

	//抄送
	//m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", "[alert]"+model)
	m.SetBody("text/html", strings.Join(info, "</br>"))
	//m.Attach("/home/Alex/lolcat.jpg")

	d := gomail.NewDialer("smtp.mxhichina.com", 25, "hm@wident.cn", "Wident321")

	// Send the email to Bob, Cora and Dan.
	for i := 0; i < 10; i++ {
		if err := d.DialAndSend(m); err != nil {
			log.Println("[util.email] send email failed ", err.Error())
			time.Sleep(time.Second * 5)
			continue
		}
		break
	}

}
