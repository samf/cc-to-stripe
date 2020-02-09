package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/samf/cc-to-stripe/assets"
	log "github.com/sirupsen/logrus"
)

type custInfo struct {
	Hostname      string `envconfig:"hostname" required:"true" desc:"hostname your customer will use"`
	Path          string `envconfig:"path" required:"true" desc:"path part of URL for customer"`
	Name          string `envconfig:"name" required:"true" desc="full name of customer"`
	StripeCust    string `envconfig:"stripe_cust" required:"true" desc:"stripe ID for customer"`
	StripePrivate string `envconfig:"stripe_private" desc:"override global private key"`
	StripePublic  string `envconfig:"stripe_public" desc:"override global public key"`
}

var (
	custMap   map[string]custInfo
	errNoCust = errors.New("no such customer entry")
)

func readCust() error {
	var ci custInfo
	cmap := make(map[string]custInfo)

	for _, cust := range config.Customers {
		err := envconfig.Process(cust, &ci)
		if err != nil {
			envconfig.Usage(cust, &ci)
			return err
		}

		if ci.StripePrivate == "" {
			ci.StripePrivate = config.StripePrivate
		}
		if ci.StripePublic == "" {
			ci.StripePublic = config.StripePublic
		}

		cmap[ci.Hostname] = ci
	}

	custMap = cmap

	return nil
}

func reqToHost(r *http.Request) (*custInfo, error) {
	host := r.Host
	if host == "localhost"+config.HTTPPort {
		host = config.LocalhostOverride
	}
	cust, ok := custMap[host]
	if !ok {
		vhost := fmt.Sprintf("%#v", host)
		log.WithFields(log.Fields{"key": vhost, "error": errNoCust}).
			Warn("no such customer")
		return nil, errNoCust
	}
	return &cust, nil
}

func custRouter() {
	ccEntryFile, err := assets.Assets.Open("entry.html")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Panic("reading entry.html")
	}
	defer ccEntryFile.Close()
	ccEntry, err := ioutil.ReadAll(ccEntryFile)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Panic("reading entry.html")
	}
	ccTemplate := template.Must(template.New("ccTemplate").
		Parse(string(ccEntry)))

	// the "/" handler either shows the main page or fails
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		cust, err := reqToHost(r)
		if err != nil {
			notFound(w, r)
			return
		}
		wanted := cust.Path
		if !strings.HasPrefix(wanted, "/") {
			wanted = "/" + wanted
		}
		if r.URL.Path != wanted {
			log.WithFields(log.Fields{
				"customer": cust.Name,
				"path":     r.URL.Path,
			}).Warn("incorrect path")
			notFound(w, r)
			return
		}

		err = ccTemplate.Execute(w, cust)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).
				Fatal("template render")
		}
	})
}

func custHostPolicy(ctx context.Context, host string) error {
	_, ok := custMap[host]

	if !ok {
		return errNoCust
	}

	return nil
}
