package bitget

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Arg struct {
	InstType string `json:"instType"`
	Channel  string `json:"channel"`
	InstID   string `json:"instId"`
}

type MarketChanInfo struct {
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

type Request struct {
	Op   string `json:"op"`
	Args []Arg  `json:"args"`
}

type Response struct {
	Event string `json:"event"`
	Arg   `json:"arg"`
}

type MarketChanPushedData struct {
	Action string `json:"action"`
	Arg    `json:"arg"`
	Data   []MarketChanInfo `json:"data"`
	Ts     int64            `json:"ts"`
}

type Channel struct {
	conn *websocket.Conn
}

func (ch *Channel) Close() error {
	return ch.conn.Close()
}

func (ch *Channel) Subscribe(ctx context.Context, req Request) (chan []byte, chan error) {
	out := make(chan []byte, 1)
	errs := make(chan error, 1)
	err := ch.conn.WriteJSON(&req)
	if err != nil {
		errs <- err

		err := ch.conn.Close()
		if err != nil {
			errs <- err
		}

		close(errs)
		close(out)

		return out, errs
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
	LOOP:
		for {
			select {
			case <-ctx.Done():
				break LOOP
			default:
				time.Sleep(time.Second * 28)
				for i := 0; i < 5; i++ {
					err := ch.conn.WriteMessage(websocket.TextMessage, []byte("ping"))
					if err == nil {
						break
					} else {
						errs <- err
						continue
					}
				}
			}
		}
	}()

	go func() {
		defer wg.Done()
	LOOP:
		for {
			select {
			case <-ctx.Done():
				break LOOP
			default:
				_, r, err := ch.conn.NextReader()
				if err != nil {
					errs <- err
					break LOOP
				}
				data, err := io.ReadAll(r)
				if err != nil {
					fmt.Println(err.Error())
					continue
				}
				out <- data
			}
		}
	}()

	go func() {
		wg.Wait()
		close(out)
		close(errs)
	}()

	return out, nil
}

func NewChannel(rBufSz int, wBufSz int, timeout time.Duration, wsURL string) (*Channel, error) {
	d := websocket.Dialer{
		ReadBufferSize:   rBufSz,
		WriteBufferSize:  wBufSz,
		HandshakeTimeout: timeout,
	}
	conn, _, err := d.Dial(wsURL, nil)
	if err != nil {
		return nil, err
	}
	ch := Channel{
		conn: conn,
	}
	return &ch, nil
}
