package main

import (
	"fmt"
	gomail "gopkg.in/mail.v2"
)

func sendEmail(fileLocation string, receiverEmail string) error {
	m:=gomail.NewMessage()

	//sender email 입력
	m.SetHeader("From", "SenderNaverEmail")

	//receiver email
	m.SetHeader("To",receiverEmail)

	//제목
	m.SetHeader("Subject","GoChatting backup messages")

	//본문
	m.SetBody("text/plain","Here is the backup message you requested.")

	//파일 첨부
	m.Attach(fileLocation)

	//네이버 id,password 입력 필수.
	d:=gomail.NewDialer("smtp.naver.com",587,"NaverID","NaverPassword")

	if err:=d.DialAndSend(m); err!=nil{
		return fmt.Errorf("failed to send mail")
	}

	return nil
}