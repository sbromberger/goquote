package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func encodeQueryParams(s []string) string {
	vals := url.Values{
		"symbols": []string{strings.Join(s, ",")},
		"types":   []string{"quote"},
		"last":    []string{"1"},
	}

	return vals.Encode()
}

func getsymb(client *http.Client, s []string) (map[string]IEXQuote, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.iextrading.com/1.0/stock/market/batch", nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = encodeQueryParams(s)
	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		errmsg := fmt.Sprintf("request returned statuscode %d (%s)", res.StatusCode, res.Status)
		return nil, errors.New(errmsg)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	iex := make(map[string]IEXQuote)

	if err := json.Unmarshal(body, &iex); err != nil {
		return nil, err
	}
	return iex, nil
}

func main() {
	var sort = flag.String("s", "+symbol", "sort by this column")
	var cert = flag.String("cafile", "", "file for SSL certificate")
	flag.Parse()

	var fieldmap = map[string]int{
		"symbol":    0,
		"sym":       0,
		"s":         0,
		"latest":    1,
		"l":         1,
		"open":      2,
		"o":         2,
		"op":        2,
		"close":     3,
		"c":         3,
		"cl":        3,
		"change":    4,
		"chg":       4,
		"ch":        4,
		"changepct": 5,
		"chp":       5,
		"chgp":      5,
		"chgpct":    5,
		"pct":       5,
		"pctchg":    5,
		"pchg":      5,
		"%":         5,
		"time":      6,
		"t":         6,
		"volume":    7,
		"vol":       7,
		"v":         7,
	}

	var sortDescending bool
	var sortField string
	if *sort != "" {
		if strings.HasPrefix(*sort, "-") {
			sortDescending = true
			sortField = string([]rune(*sort)[1:])
		} else if strings.HasPrefix(*sort, "+") {
			sortDescending = false
			sortField = string([]rune(*sort)[1:])
		} else {
			sortDescending = false
			sortField = *sort
		}
	}

	symbs := flag.Args()
	sortCol := fieldmap[sortField]

	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Read in the cert file
	if *cert != "" {
		certs, err := ioutil.ReadFile(*cert)
		if err != nil {
			log.Fatalf("Failed to append %q to RootCAs: %v", *cert, err)
		}

		// Append our cert to the system pool
		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			log.Println("No certs appended, using system certs only")
		}
	}
	config := &tls.Config{
		RootCAs: rootCAs,
	}
	tr := &http.Transport{TLSClientConfig: config}
	client := http.Client{Timeout: time.Second * 4, Transport: tr}

	iex, err := getsymb(&client, symbs)
	if err != nil {
		log.Fatal(err)
	}
	render(iex, sortCol, sortDescending)
}
