package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

// IEXQuote holds data from iextrading.com.
type IEXQuote struct {
	Q Quote `json:"quote"`
}

// Quote holds actual quote data. This is necessary for unmarshaling.
type Quote struct {
	Symbol    string  `json:"symbol"`
	Open      float64 `json:"open"`
	Close     float64 `json:"close"`
	Latest    float64 `json:"latestPrice"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"changePercent"`
	AsOf      int64   `json:"latestUpdate"`
	Volume    int     `json:"latestVolume"`
}

func floatToString(f float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func colorizeFloatToString(f float64) string {
	// to convert a float number to a string
	switch {
	case f < 0:
		return color.RedString(floatToString(f))
	case f > 0:
		return color.GreenString(floatToString(f))
	default:
		return floatToString(f)
	}
}

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

func render(iex map[string]IEXQuote) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Sym", "Latest", "Open", "Close", "Chg", "%Chg", "VolM", "Time"})
	r := tablewriter.ALIGN_RIGHT
	table.SetColumnAlignment([]int{r, r, r, r, r, r, r, r})
	var keys []string
	for k := range iex {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := iex[k]
		ts := time.Unix(0, v.Q.AsOf*1000000)
		tz, _ := ts.Zone()
		tsString := ts.Format("2006-01-02 15:04:05 ") + tz
		row := []string{
			v.Q.Symbol,
			floatToString(v.Q.Latest),
			floatToString(v.Q.Open),
			floatToString(v.Q.Close),
			colorizeFloatToString(v.Q.Change),
			colorizeFloatToString(v.Q.ChangePct),
			floatToString(float64(v.Q.Volume) / 1000000),
			tsString,
		}
		table.Append(row)
	}
	table.Render()
}

func main() {
	client := http.Client{Timeout: time.Second * 4}
	symbs := os.Args[1:]
	iex, err := getsymb(&client, symbs)
	if err != nil {
		log.Fatal(err)
	}
	render(iex)
}
