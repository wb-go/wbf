package rabbitmq

import "errors"

var (
	// ErrClientClosed возвращается при попытке выполнить операцию над уже закрытым клиентом RabbitMQ.
	ErrClientClosed = errors.New("rabbitmq client closed")
	// ErrChannelLost возвращается, когда канал становится непригодным из-за разрыва соединения.
	ErrChannelLost = errors.New("channel lost due to connection drop")
	// ErrMissingURL возвращается, если URL для подключения к RabbitMQ не указан.
	ErrMissingURL = errors.New("rabbitmq URL is required")
	// ErrChannelClosedUnexpectedly возвращается, когда канал доставки сообщений
	// был закрыт неожиданно (например, из-за потери соединения).
	ErrChannelClosedUnexpectedly = errors.New("message channel closed unexpectedly")
)
