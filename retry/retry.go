// Package retry предоставляет функциональность повторных попыток с настраиваемыми стратегиями.
package retry

import (
	"context"
	"time"
)

// Strategy определяет параметры поведения повторных попыток.
type Strategy struct {
	Attempts int           // Количество попыток.
	Delay    time.Duration // Начальная задержка между попытками.
	Backoff  float64       // Множитель для увеличения задержки.
}

// Do выполняет функцию с заданной стратегией повторных попыток.
func Do(fn func() error, strategy Strategy) error {
	delay := strategy.Delay
	var err error
	for i := 0; i < strategy.Attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		time.Sleep(delay)
		delay = time.Duration(float64(delay) * strategy.Backoff)
	}
	return err
}

// DoContext выполняет функцию с заданной стратегией повторных попыток только с контекстом и завершением при gracefully shutdown
func DoContext(ctx context.Context, strategy Strategy, fn func(context.Context) error) error {
	delay := strategy.Delay
	var err error
	for i := 0; i < strategy.Attempts; i++ {
		err = fn(ctx)
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		delay = time.Duration(float64(delay) * strategy.Backoff)
	}
	return err
}
