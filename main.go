//go:generate go run -tags=dev gendata.go

package main

import (
	"crypto/tls"
	"flag"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/samf/cc-to-stripe/assets"
	log "github.com/sirupsen/logrus"

	"golang.org/x/crypto/acme/autocert"
)

var insecure = flag.Bool("i", false, "turn off security (http only)")
var httpPort = flag.String("p", ":80", "port for http")
var httpsPort = flag.String("P", ":443", "port for https")
var cachedir = flag.String("C", "/autocert", "directory for autocert cache")
var email = flag.String("email", "sam.falkner@gmail.com",
	"email for letsencrypt")
var localhost = flag.String("l",
	"localhost", "host to use if localhost detected")

var notFound, err500 http.HandlerFunc

func main() {
	log.WithFields(log.Fields{"argv": os.Args}).Info("start")
	flag.Parse()

	if err := readCust(); err != nil {
		log.WithFields(log.Fields{"error": err}).
			Fatal("readCust failed")
		os.Exit(1)
	}

	mainRouter()

	if *insecure {
		log.WithFields(log.Fields{"addr": *httpPort}).Info("insecure")
		err := http.ListenAndServe(*httpPort, nil)
		log.WithFields(log.Fields{"error": err}).Fatal("http returned")
	}

	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(*cachedir),
		HostPolicy: custHostPolicy,
		Email:      *email,
	}

	go func() {
		handler := m.HTTPHandler(nil)
		log.WithFields(log.Fields{"port": *httpPort}).
			Info("redirector starting")
		err := http.ListenAndServe(*httpPort, handler)
		log.WithFields(log.Fields{"error": err}).Fatal("http redirect")
	}()

	s := &http.Server{
		Addr: *httpsPort,
		TLSConfig: &tls.Config{
			GetCertificate: m.GetCertificate,
		},
	}
	log.WithFields(log.Fields{"port": *httpsPort}).Info("secure")
	err := s.ListenAndServeTLS("", "")
	log.WithFields(log.Fields{"error": err}).Fatal("https returned")
}

func mainRouter() {
	notFoundFile, err := assets.Assets.Open("404.html")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Panic("opening 404.html")
	}
	defer notFoundFile.Close()
	notFoundHTML, err := ioutil.ReadAll(notFoundFile)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("reading 404.html")
		notFoundHTML = []byte("file not found")
	}
	notFound = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write(notFoundHTML)
	}

	html500file, err := assets.Assets.Open("500.html")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Panic("opening 500.html")
	}
	defer html500file.Close()
	html500, err := ioutil.ReadAll(html500file)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("reading 500.html")
		html500 = []byte("internal error")
	}
	err500 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write(html500)
	}

	mainCssFile, err := assets.Assets.Open("main.css")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Panic("opening main.css")
	}
	defer mainCssFile.Close()
	mainCSS, err := ioutil.ReadAll(mainCssFile)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("reading main.css")
	}
	http.HandleFunc("/main.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Write(mainCSS)
	})

	faviconFile, err := assets.Assets.Open("favicon-32x32.png")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Panic("opening favicon")
	}
	defer faviconFile.Close()
	favicon, err := ioutil.ReadAll(faviconFile)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Panic("reading favicon")
	}
	http.HandleFunc("/favicon-32x32.png",
		func(w http.ResponseWriter, r *http.Request) {
			// Let the http library figure out the file type.
			// That way people can use their own favicons in any format.
			w.Write(favicon)
		})

	// the top level "/" handler is in custRouter
	custRouter()
	stripeRouter()
}
