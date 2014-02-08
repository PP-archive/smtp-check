package main

import (
	"fmt"
	"net"
	"net/smtp"
	"os"
	"time"
	/*"net/textproto"*/
)

func main() {
	fmt.Printf("Start!\n")

//	mxServers, err := net.LookupMX("yandex.ru")
//
//	if(err != nil) {
//		fmt.Println("Error:", err)
//		os.Exit(1)
//	}
//
//	for key, mxServer := range mxServers {
//		fmt.Println((key+1), ") OK:", mxServer.Host, mxServer.Pref)
//	}
//
//	os.Exit(1)

	smtpHost := "mx.yandex.ru"
	smtpPort := "25"

	// to support timeout
	timeout, _ := time.ParseDuration("10s")
	Conn, err := net.DialTimeout("tcp", smtpHost+":"+smtpPort, timeout)

	Client, err := smtp.NewClient(Conn, smtpHost)

	if(err != nil) {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Connected to "+smtpHost+":"+smtpPort)

	email1 := "ka-zaido@yandex.ru"
	//fmt.Println("1] Checking "+email1)
	//err = Client.Verify(email1)

	err = Client.Hello("hi")

	if(err != nil) {
		fmt.Println(err)
	}
	err = Client.Mail("pavel@kredito.de")

	if(err != nil) {
		fmt.Println(err)
	}
	err = Client.Rcpt(email1)

	if(err != nil) {
		fmt.Println(err)
		fmt.Println(email1+" is not verified")
	} else {
		fmt.Println(email1+" is verified")
	}

	Client.Quit()
	Client.Close()

	// 2nd mail
	Client, _ = smtp.Dial(smtpHost+":"+smtpPort)

	email2 := "ka-z444aido@yandex.ru"

	err = Client.Hello("hi")

	if(err != nil) {
		fmt.Println(err)
	}
	err = Client.Mail("pavel@kredito.de")

	if(err != nil) {
		fmt.Println(err)
	}
	err = Client.Rcpt(email2)

	if(err != nil) {
		fmt.Println(err)
		fmt.Println(email2+" is not verified")
	} else {
		fmt.Println(email2+" is verified")
	}

	Client.Quit()
	Client.Close()

	fmt.Println("Check complete")
}
