package main

import (
	"fmt"
	"net/smtp"
	"os"
	/*"net/textproto"*/
)

func main() {
	fmt.Printf("Hello world!\n")

	smtpHost := "smtp.yandex.ru"
	smtpPort := "25"

	Client, err := smtp.Dial(smtpHost+":"+smtpPort)

	if(err != nil) {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Connected to "+smtpHost+":"+smtpPort)

	email1 := "ka-zaido@yandex.ru"
	fmt.Println("1] Checking "+email1)
	err = Client.Verify(email1)

	if(err != nil) {
		fmt.Println(err)
		fmt.Println(email1+" is not verified")
	} else {
		fmt.Println(email1+" is verified")
	}

	email2 := "ka-z444aido@yandex.ru"
	fmt.Println("2] Checking "+email2)
	err = Client.Verify(email2)

	if(err != nil) {
		fmt.Println(err)
		fmt.Println(email2+" is not verified")
	} else {
		fmt.Println(email2+" is verified")
	}

	fmt.Println("Check complete")
}
