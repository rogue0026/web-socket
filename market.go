package main

import (
	"fmt"
	"strconv"
)

type Arg struct {
	InstType string `json:"instType"`
	Channel  string `json:"channel"`
	InstID   string `json:"instId"`
}

type Info struct {
	InstID       string `json:"instId"`
	LastPr       string `json:"lastPr"`
	Open24H      string `json:"open24h"`
	High24H      string `json:"high24h"`
	Low24H       string `json:"low24h"`
	Change24H    string `json:"change24h"`
	BidPr        string `json:"bidPr"`
	AskPr        string `json:"askPr"`
	BidSz        string `json:"bidSz"`
	AskSz        string `json:"askSz"`
	BaseVolume   string `json:"baseVolume"`
	QuoteVolume  string `json:"quoteVolume"`
	OpenUtc      string `json:"openUtc"`
	ChangeUtc24H string `json:"changeUtc24h"`
	Ts           string `json:"ts"`
}

type MarketChanRequest struct {
	Op   string `json:"op"`
	Args []Arg  `json:"args"`
}

type MarketChanResponse struct {
	Event string `json:"event"`
	Arg   `json:"arg"`
}

type PushedData struct {
	Action string `json:"action"`
	Arg    `json:"arg"`
	Data   []Info `json:"data"`
	Ts     int64  `json:"ts"`
}

func (pd PushedData) String() string {
	if len(pd.Data) == 0 {
		return ""
	} else {
		lastPrice, err := strconv.ParseFloat(pd.Data[0].LastPr, 64)
		if err != nil {
			// todo
		}
		bidPr, err := strconv.ParseFloat(pd.Data[0].BidPr, 64)
		if err != nil {
			// todo
		}
		askPr, err := strconv.ParseFloat(pd.Data[0].AskPr, 64)
		if err != nil {
			// todo
		}
		spread := (askPr - bidPr) / askPr * 100.0
		return fmt.Sprintf("pair: %s\tcurrent price: %f\tspread: %f", pd.InstID, lastPrice, spread)
	}
}
