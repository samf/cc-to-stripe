package main

import (
	"io/ioutil"
	"net/http"

	"github.com/samf/cc-to-stripe/assets"
	log "github.com/sirupsen/logrus"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/client"
)

func stripeRouter() {
	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		cust, err := reqToHost(r)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).
				Warn("call reqToHost")
			err500(w, r)
			return
		}
		sc := new(client.API)
		sc.Init(cust.StripePrivate, nil)

		err = r.ParseForm()
		if err != nil {
			log.WithFields(log.Fields{"error": err}).
				Warn("ParseForm")
			return
		}

		token := r.PostFormValue("stripeToken")
		if token == "" {
			log.Warn("stripeToken empty")
			err500(w, r)
			return
		}
		log.WithFields(log.Fields{"token": token}).Info("stripeToken")
		params := &stripe.CustomerParams{
			Source: &stripe.SourceParams{
				Token: &token,
			},
		}

		_, err = sc.Customers.Update(cust.StripeCust, params)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).
				Warn("update customers failed")
			err500(w, r)
			return
		}

		http.Redirect(w, r, "/success", http.StatusTemporaryRedirect)
	})

	successHtmlFile, err := assets.Assets.Open("success.html")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("opening success.html")
	}
	defer successHtmlFile.Close()
	successHTML, err := ioutil.ReadAll(successHtmlFile)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("reading success.html")
		successHTML = []byte("you have successfully updated your card")
	}
	http.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(successHTML)
	})
}
