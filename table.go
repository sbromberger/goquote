package main

import (
	"math"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

// change that will cause bolding of price fields and
// an exclamation mark at the beginning of each row
const bigmove = 0.05

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
	default:
		c = color.New(color.FgWhite)
		c.DisableColor()
		defer c.EnableColor()
	}

	if bold {
		c.Add(color.Bold)
	}
	return c.Sprint(s)
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

func render(iex map[string]IEXQuote, sortCol int, sortDescending bool) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"!", "Sym", "Latest", "Open", "Close", "Chg", "%Chg", "VolM", "Time"})
	r := tablewriter.ALIGN_RIGHT
	table.SetColumnAlignment([]int{r, r, r, r, r, r, r, r, r})

	var quoteSlice []Quote
	for _, v := range iex {
		quoteSlice = append(quoteSlice, v.Q)
	}

	sortFn := sortFns[sortCol]
	sortFn(quoteSlice)
	if sortDescending {
		var qsrev []Quote
		for i := len(quoteSlice) - 1; i >= 0; i-- {
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
			colorizeftoa(v.ChangePct*100.0, alertbigmove),
			ftoa(float64(v.Volume) / 1000000),
			tsString,
		}
		table.Append(row)
	}
	table.Render()
}
