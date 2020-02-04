package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/samf/cc-to-stripe/assets"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type (
	custInfo struct {
		Hostname      string `yaml:"hostname"`
		Path          string `yaml:"path"`
		Name          string `yaml:"name"`
		StripeCust    string `yaml:"cust_id"`
		StripePrivate string `yaml:"stripe_secret"`
		StripePublic  string `yaml:"stripe_public"`
	}
)

var (
	custMap   map[string]custInfo
	errNoCust = errors.New("no such customer entry")
)

func readCust() error {
	var custList []custInfo
	cmap := make(map[string]custInfo)
	rawconfigFile, err := assets.Assets.Open("config.yaml")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("opening config")
		return err
	}
	defer rawconfigFile.Close()
	rawconfig, err := ioutil.ReadAll(rawconfigFile)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("reading config")
		return err
	}

	err = yaml.Unmarshal(rawconfig, &custList)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("parsing config")
		return err
	}

	for _, cust := range custList {
		cmap[cust.Hostname] = cust
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
