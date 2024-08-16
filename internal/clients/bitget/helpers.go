package bitget

import "fmt"

func CreateSubscribeRequest(cryptocurrency string) Request {
	req := Request{
		Op: "subscribe",
		Args: []Arg{
			{
				InstType: "SPOT",
				Channel:  "ticker",
				InstID:   fmt.Sprintf("%sUSDT", cryptocurrency),
			},
		},
	}

	return req
}
