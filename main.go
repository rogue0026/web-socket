package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var (
	wsEndpt = "wss://ws.bitget.com/v2/ws/public"
)

type MarketChannel struct {
	conn *websocket.Conn
}

func NewMarketChannel(url string, headers http.Header, wrBufSize int, rBufSize int) (MarketChannel, error) {
	d := websocket.Dialer{
		WriteBufferSize:  wrBufSize,
		ReadBufferSize:   rBufSize,
		HandshakeTimeout: time.Second * 5,
	}
	c, _, err := d.Dial(url, headers)
	if err != nil {
		return MarketChannel{}, err
	}
	mc := MarketChannel{
		conn: c,
	}
	return mc, nil
}

func (mc *MarketChannel) Subscribe(ctx context.Context, req MarketChanRequest) (chan []byte, chan error) {
	out := make(chan []byte, 1)
	errs := make(chan error, 1)

	err := mc.conn.WriteJSON(&req)
	if err != nil {
		close(out)
		errs <- err
		return nil, errs
	}

	go func() {
		for {
			time.Sleep(time.Second * 30)
			select {
			case <-ctx.Done():
				return
			default:
				for i := 0; i < 5; i++ {
					err := mc.conn.WriteMessage(websocket.TextMessage, []byte("ping"))
					if err != nil {
						errs <- err
						continue
					} else {
						break
					}
				}
			}
		}
	}()

	go func() {
	LOOP:
		for {
			select {
			case <-ctx.Done():
				close(errs)
				close(out)
				return
			default:
				time.Sleep(time.Millisecond * 100)
				_, r, err := mc.conn.NextReader()
				if err != nil {
					fmt.Println(err.Error())
					close(out)
					close(errs)
					break LOOP
				}
				data, err := io.ReadAll(r)
				if err != nil {
					errs <- err
				}
				out <- data
			}
		}
	}()

	return out, errs
}

func main() {
	mc, err := NewMarketChannel(wsEndpt, nil, 256, 256)
	if err != nil {
		log.Fatalln(err.Error())
	}
	req := MarketChanRequest{
		Op: "subscribe",
		Args: []Arg{
			{
				InstType: "SPOT",
				Channel:  "ticker",
				InstID:   "BTCUSDT",
			},
		},
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	results, errs := mc.Subscribe(ctx, req)

MAIN_LOOP:
	for {
		select {
		case err := <-errs:
			fmt.Printf("%s", err.Error())
			fmt.Printf("\r")
		case js := <-results:
			pd := PushedData{}
			err := json.Unmarshal(js, &pd)
			if err != nil {
				fmt.Printf("%s", err.Error())
				fmt.Printf("\r")
			} else {
				fmt.Printf("%s", pd)
				fmt.Printf("\r")
			}
		case <-stop:
			break MAIN_LOOP
		}
	}

	cancel()
}
