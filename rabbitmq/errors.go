package rabbitmq

import "errors"

var (
	ErrClientClosed = errors.New("rabbitmq client closed")
	ErrChannelLost  = errors.New("channel lost due to connection drop")
)
