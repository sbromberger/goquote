package main

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
