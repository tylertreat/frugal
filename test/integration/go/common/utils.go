package common

import "github.com/nats-io/nats"

func getNatsConn() (*nats.Conn, error) {
	addr := nats.DefaultURL
	natsOptions := nats.DefaultOptions
	natsOptions.Servers = []string{addr}
	natsOptions.Secure = false
	conn, err := natsOptions.Connect()
	if err != nil {
		return nil, err
	}

	return conn, nil
}
