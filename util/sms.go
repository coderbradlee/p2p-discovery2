package util

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

//
var smsSign = "【91Pool】"

func smsSendAction(mobile []string, content string) {
	for _, v := range mobile {
		str := "username=xujing&password=xjMall!@#136xj&mobile=" + v + "&content=" + content + smsSign + "&xh="
		var result string
		var err error
		for i := 0; i < 10; i++ {
			result, err = sendMessage("http://114.215.196.145/sendSmsApi", str)
			if err != nil {
				log.Println("次发送短信失败，五秒钟后重新发送", i+1, err, result)
				time.Sleep(time.Second * 5)
				continue
			}
			break
		}
		log.Println("发送短信:" + v + "[" + content + "],返回结果:" + result)
	}
}

func sendMessage(uri, reqBody string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		return "Create request error", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("cache-control", "no-cache")

	resp, err := client.Do(req)
	if err != nil {
		return "Send message error", err
	}
	defer resp.Body.Close()
	respStr, _ := ioutil.ReadAll(resp.Body)
	return string(respStr), nil
}

func SmsSend(content string) {
	mobiles, err := LoadMobile()
	if err != nil {
		log.Println("sms", "LoadMobile err", err, content)
		return
	}
	smsSendAction(mobiles, content)
}

type MobileToml struct {
	Mobiles []string
}

func LoadMobile() (tels []string, err error) {

	mt := &MobileToml{}

	f, err := os.Open("contacts.toml")
	if err != nil {
		return
	}
	defer f.Close()

	if _, err = toml.DecodeReader(f, mt); err != nil {
		return
	}

	tels = mt.Mobiles

	return
}
