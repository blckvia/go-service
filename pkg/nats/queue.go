package nats

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
)

type NatsService interface {
	Publish(ctx context.Context, subject string, data interface{}) error
	Subscribe(ctx context.Context, subject string, handler func(msg *nats.Msg)) error
}

type NatsClient struct {
	conn *nats.Conn
}

func NewNatsClient(conn *nats.Conn) *NatsClient {
	return &NatsClient{conn: conn}
}

func (n *NatsClient) Publish(ctx context.Context, subject string, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return n.conn.Publish(subject, dataBytes)
}

func (n *NatsClient) Subscribe(ctx context.Context, subject string, handler func(msg *nats.Msg)) error {
	_, err := n.conn.Subscribe(subject, func(msg *nats.Msg) {
		handler(msg)
	})
	return err
}
