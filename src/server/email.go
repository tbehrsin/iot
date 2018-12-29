package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"gateway/errors"
	"log"
	"net/http"
	"os"

	mg "github.com/mailgun/mailgun-go"
)

var mailgunDomain = os.Getenv("MAILGUN_DOMAIN")
var mailgunAPIKey = os.Getenv("MAILGUN_API_KEY")
var mailgunSender = fmt.Sprintf("Behrsin IoT <noreply@%s>", mailgunDomain)
var mailgun = mg.NewMailgun(mailgunDomain, mailgunAPIKey)

type CreateEmailTokenRequest struct {
	Email string
	Token string
}

func CreateEmailToken(w http.ResponseWriter, r *http.Request) {
	var req CreateEmailTokenRequest

	if b, err := ioutil.ReadAll(r.Body); err != nil {
		errors.NewBadRequest(err).Println().Write(w)
	} else if err := json.Unmarshal(b, &req); err != nil {
		log.Println(string(b))
		errors.NewBadRequest(err).Println().Write(w)
	} else {
		sendMessage(
			"Create your Behrsin IoT account...",
			fmt.Sprintf(`
      Click the following link to create your Behrsin IoT account using the online dashboard:

      https://iot.behrsin.com/#token=%s
      `, req.Token),
			req.Email,
		)
		w.Write([]byte("{\"body\":null}"))
	}
}

func sendMessage(subject string, body string, recipient string) {
	log.Println(subject, body, recipient, mailgunDomain, mailgunAPIKey, mailgunSender)
	message := mailgun.NewMessage(mailgunSender, subject, body, recipient)
	if resp, id, err := mailgun.Send(message); err != nil {
		log.Println(err)
	} else {
		fmt.Printf("sent mail %s -> %s\n", id, resp)
	}
}
