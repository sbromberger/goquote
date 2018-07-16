package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
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

// percent change that will cause bolding of price fields and
// an exclamation mark at the beginning of each row
const bigmove = 5.00

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

func ftoa(f float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func colorizeftoa(f float64, bold bool) string {
	// to convert a float number to a string
	var c *color.Color
	s := ftoa(f)
	switch {
	case f < 0:
		c = color.New(color.FgRed)
	case f > 0:
		c = color.New(color.FgGreen)
	}

	if bold {
		c.Add(color.Bold)
	}
	return c.Sprint(s)
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

func render(iex map[string]IEXQuote, sortCol int, sortDescending bool) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"!", "Sym", "Latest", "Open", "Close", "Chg", "%Chg", "VolM", "Time"})
	r := tablewriter.ALIGN_RIGHT
	table.SetColumnAlignment([]int{r, r, r, r, r, r, r, r, r})
	var keys []string

	for k := range iex {
		keys = append(keys, k)
	}

	var quoteSlice []Quote
	for _, v := range iex {
		quoteSlice = append(quoteSlice, v.Q)
	}

	sortFn := sortFns[sortCol]
	sortFn(quoteSlice)
	if sortDescending {
		var qsrev []Quote
		for i := len(quoteSlice) - 1; i >= 0; i-- {
			fmt.Println("i = ", i)
			qsrev = append(qsrev, quoteSlice[i])
		}
		quoteSlice = qsrev
	}

	for _, v := range quoteSlice {
		ts := time.Unix(0, v.AsOf*1000000)
		tz, _ := ts.Zone()
		tsString := ts.Format("01-02 15:04:05 ") + tz
		alertbigmove := math.Abs(v.ChangePct) > bigmove
		rowalert := " "
		if alertbigmove {
			rowalert = "!"
		}
		row := []string{
			rowalert,
			v.Symbol,
			ftoa(v.Latest),
			ftoa(v.Open),
			ftoa(v.Close),
			colorizeftoa(v.Change, alertbigmove),
			colorizeftoa(v.ChangePct, alertbigmove),
			ftoa(float64(v.Volume) / 1000000),
			tsString,
		}
		table.Append(row)
	}
	table.Render()
}

func sortBySymbol(quotes []Quote) {
	sort.Slice(quotes, func(i, j int) bool { return quotes[i].Symbol < quotes[j].Symbol })
}

func sortByLatest(quotes []Quote) {
	sort.Slice(quotes, func(i, j int) bool { return quotes[i].Latest < quotes[j].Latest })
}

func sortByOpen(quotes []Quote) {
	sort.Slice(quotes, func(i, j int) bool { return quotes[i].Open < quotes[j].Open })
}

func sortByClose(quotes []Quote) {
	sort.Slice(quotes, func(i, j int) bool { return quotes[i].Close < quotes[j].Close })
}

func sortByChange(quotes []Quote) {
	sort.Slice(quotes, func(i, j int) bool { return quotes[i].Change < quotes[j].Change })
}

func sortByChangePct(quotes []Quote) {
	sort.Slice(quotes, func(i, j int) bool { return quotes[i].ChangePct < quotes[j].ChangePct })
}

func sortByTime(quotes []Quote) {
	sort.Slice(quotes, func(i, j int) bool { return quotes[i].AsOf < quotes[j].AsOf })
}

func sortByVol(quotes []Quote) {
	sort.Slice(quotes, func(i, j int) bool { return quotes[i].Volume < quotes[j].Volume })
}

var sortFns = [...]func([]Quote){
	sortBySymbol,
	sortByLatest,
	sortByOpen,
	sortByClose,
	sortByChange,
	sortByChangePct,
	sortByTime,
	sortByVol,
}

func main() {
	var sort = flag.String("s", "+symbol", "sort by this column")
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
		"time":      6,
		"t":         6,
		"volume":    7,
		"vol":       7,
		"v":         7,
	}

	var sortDescending bool
	var sortField string
	if strings.HasPrefix(*sort, "+") {
		sortDescending = false
		sortField = string([]rune(*sort)[1:])
	} else if strings.HasPrefix(*sort, "-") {
		sortDescending = true
		sortField = string([]rune(*sort)[1:])
	} else {
		sortDescending = true
	}

	sortCol := fieldmap[sortField]

	// fmt.Println("sortCol = ", sortCol)
	// fmt.Println("sortDescending = ", sortDescending)

	client := http.Client{Timeout: time.Second * 4}
	symbs := os.Args[1:]
	iex, err := getsymb(&client, symbs)
	if err != nil {
		log.Fatal(err)
	}
	render(iex, sortCol, sortDescending)
}
