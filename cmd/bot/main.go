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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// for gracefull shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	req := bitget.CreateSubscribeRequest("LTC")

	// creating new market channel
	markCh, err := bitget.NewChannel(256, 256, time.Second*5, webSocketDomain)
	if err != nil {
		log.Fatal(err.Error())
	}
	data, errs := markCh.Subscribe(ctx, req)

	// start receiving data
	go func() {
		for err := range errs {
			wr.WriteToConsole(err.Error())
		}
	}()

	curPriceCh := make(chan float64, 1)
	go func() {
		defer close(curPriceCh)
		for d := range data {
			pushData := bitget.MarketChanPushedData{}
			if err := json.Unmarshal(d, &pushData); err != nil {
				continue
			} else {
				if len(pushData.Data) > 0 {
					curPrice, _ := strconv.ParseFloat(pushData.Data[0].LastPr, 64)
					curPriceCh <- curPrice
					bid, _ := strconv.ParseFloat(pushData.Data[0].BidPr, 64)
					ask, _ := strconv.ParseFloat(pushData.Data[0].AskPr, 64)
					spread := ask - bid
					wr.WriteToConsole(fmt.Sprintf("%s LTC: %.6f\t Spread: %.6f", time.Now().Format("02.01.2006 15:04:05:"), curPrice, spread))
				}
			}
		}
	}()

	// collecting price fluctuations
	tckr := time.NewTicker(time.Minute * 4)

	statFile, err := os.Create("./tmp/statistics.txt")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer statFile.Close()
	go func() {
		var minPr float64
		var maxPr float64
		for {
			select {
			case curPrice := <-curPriceCh:
				if minPr == 0 {
					minPr = curPrice
				}
				if curPrice < minPr {
					minPr = curPrice
				}
				if curPrice > maxPr {
					maxPr = curPrice
				}
			case <-tckr.C:
				statFile.WriteString(fmt.Sprintf("%s | min price = %.6f max price = %.6f\n", time.Now().Format("02.06.2006 15:04:05"), minPr, maxPr))
				minPr = 0
				maxPr = 0
			default:
				continue
			}
		}
	}()
	<-stop
}
