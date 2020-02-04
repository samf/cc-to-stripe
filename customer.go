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

type (
	appInfo struct {
		Customers []string `envconfig:"customers"`
	}

	custInfo struct {
		Hostname      string `envconfig:"hostname"`
		Path          string `envconfig:"path"`
		Name          string `envconfig:"name"`
		StripeCust    string `envconfig:"stripe_cust"`
		StripePrivate string `envconfig:"stripe_private"`
		StripePublic  string `envconfig:"stripe_public"`
	}
)

var (
	custMap   map[string]custInfo
	errNoCust = errors.New("no such customer entry")
)

func readCust() error {
	var (
		ai appInfo
		ci custInfo
	)
	cmap := make(map[string]custInfo)

	err := envconfig.Process("ccs", &ai)
	if err != nil {
		return err
	}

	for _, cust := range ai.Customers {
		err := envconfig.Process(cust, &ci)
		if err != nil {
			return err
		}

		cmap[ci.Hostname] = ci
	}

	custMap = cmap

	return nil
}

func reqToHost(r *http.Request) (*custInfo, error) {
	host := r.Host
	if host == "localhost"+*httpPort {
		host = *localhost
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
