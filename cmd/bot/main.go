package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/rogue0026/web-socket/internal/clients/bitget"
)

var (
	webSocketDomain = "wss://ws.bitget.com/v2/ws/public"
)

type StdWriter struct {
	mu  *sync.Mutex
	out io.Writer
}

func (sw *StdWriter) WriteToConsole(data string) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.out.Write([]byte("\r"))
	sw.out.Write([]byte(data))
}

func main() {
	wr := StdWriter{
		mu:  &sync.Mutex{},
		out: os.Stdout,
	}
	markCh, err := bitget.NewChannel(256, 256, time.Second*5, webSocketDomain)
	if err != nil {
		log.Fatal(err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	ethReq := bitget.Request{
		Op: "subscribe",
		Args: []bitget.Arg{
			{
				InstType: "SPOT",
				Channel:  "ticker",
				InstID:   "ETHUSDT",
			},
		},
	}
	data, errs := markCh.Subscribe(ctx, ethReq)
	go func() {
		for err := range errs {
			wr.WriteToConsole(err.Error())
		}
	}()

	times, err := os.Create("times.txt")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer func() {
		err := times.Sync()
		if err != nil {
			fmt.Println(err.Error())
		}
		if err = times.Close(); err != nil {
			fmt.Println(err.Error())
		}
	}()
	go func() {
		for d := range data {
			start := time.Now()
			pushData := bitget.MarketChanPushedData{}
			if err := json.Unmarshal(d, &pushData); err != nil {
				continue
			} else {
				if len(pushData.Data) > 0 {
					curPrice, _ := strconv.ParseFloat(pushData.Data[0].LastPr, 64)
					bid, _ := strconv.ParseFloat(pushData.Data[0].BidPr, 64)
					ask, _ := strconv.ParseFloat(pushData.Data[0].AskPr, 64)
					spread := bid - ask
					wr.WriteToConsole(fmt.Sprintf("ETH price in USDT: %.2f\t Spread: %.2f", curPrice, spread))
				}
			}
			dur := time.Since(start)
			times.WriteString(strconv.FormatInt(dur.Nanoseconds(), 10) + string('\n'))
		}
	}()
	<-stop
}
