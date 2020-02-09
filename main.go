//go:generate go run -tags=dev gendata.go

package main

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/samf/cc-to-stripe/assets"
	log "github.com/sirupsen/logrus"

	"golang.org/x/crypto/acme/autocert"
)

type appConfig struct {
	Customers     []string `envconfig:"customers" required:"true" desc:"short names for your customers"`
	StripePrivate string   `envconfig:"stripe_private" required:"true" desc:"stripe private key"`
	StripePublic  string   `envconfig:"stripe_public" required:"true" desc:"stripe public key"`

	HTTPPort  string `envconfig:"http_port" default:":80"`
	HTTPSPort string `envconfig:"https_port" default:":443"`
	CacheDir  string `envconfig:"CACHEDIR" default:"/autocert" desc:"directory where certificates are cached"`
	Email     string `required:"true" desc:"needed for automatic certificate fetching"`

	HttpOnly          bool   `split_words:"true" desc:"only use http (for development only)"`
	LocalhostOverride string `split_words:"true" desc:"map localhost to a hostname (for dev only)"`
}

var config appConfig

var notFound, err500 http.HandlerFunc

func main() {
	log.WithFields(log.Fields{"argv": os.Args}).Info("start")

	if err := envconfig.Process("ccs", &config); err != nil {
		envconfig.Usage("css", &config)
		log.WithFields(log.Fields{"error": err}).
			Fatal("initial envconfig failed")
		os.Exit(1)
	}

	if err := readCust(); err != nil {
		log.WithFields(log.Fields{"error": err}).
			Fatal("readCust failed")
		os.Exit(1)
	}

	mainRouter()

	if config.HttpOnly {
		log.WithFields(log.Fields{"addr": config.HTTPPort}).Info("insecure")
		err := http.ListenAndServe(config.HTTPPort, nil)
		log.WithFields(log.Fields{"error": err}).Fatal("http returned")
	}

	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(config.CacheDir),
		HostPolicy: custHostPolicy,
		Email:      config.Email,
	}

	go func() {
		handler := m.HTTPHandler(nil)
		log.WithFields(log.Fields{"port": config.HTTPPort}).
			Info("redirector starting")
		err := http.ListenAndServe(config.HTTPPort, handler)
		log.WithFields(log.Fields{"error": err}).Fatal("http redirect")
	}()

	s := &http.Server{
		Addr: config.HTTPSPort,
		TLSConfig: &tls.Config{
			GetCertificate: m.GetCertificate,
		},
	}
	log.WithFields(log.Fields{"port": config.HTTPSPort}).Info("secure")
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
