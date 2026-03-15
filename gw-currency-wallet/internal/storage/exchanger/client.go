package exchanger

import (
	"context"
	"math"

	exchangev1 "github.com/yaoguao/gw/protos/gen/go/exchange"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const RateScale = 100

type Client struct {
	conn   *grpc.ClientConn
	client exchangev1.ExchangeServiceClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := exchangev1.NewExchangeServiceClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) GetExchangeRateForCurrency(ctx context.Context, from, to string) (int64, error) {

	resp, err := c.client.GetExchangeRateForCurrency(ctx, &exchangev1.CurrencyRequest{
		FromCurrency: from,
		ToCurrency:   to,
	})

	if err != nil {
		return 0, err
	}

	rate := int64(math.Round(resp.Rate * RateScale))

	return rate, nil
}

func (c *Client) GetExchangeRates(ctx context.Context, base string) (map[string]float64, error) {

	resp, err := c.client.GetExchangeRates(ctx, &exchangev1.ExchangeRateRequest{
		Base: base,
	})

	if err != nil {
		return nil, err
	}

	return resp.Rates, nil
}
