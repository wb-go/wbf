package retry

import (
	"time"
)

type Strategy struct {
	Attempts int
	Delay    time.Duration
	Backoff  float64 // множитель для увеличения задержки
}

func Do(fn func() error, strat Strategy) error {
	delay := strat.Delay
	var err error
	for i := 0; i < strat.Attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		time.Sleep(delay)
		delay = time.Duration(float64(delay) * strat.Backoff)
	}
	return err
}
