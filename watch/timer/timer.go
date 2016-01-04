package timer

import (
	"fmt"
	"sync/atomic"
	"time"
)

type IntervalWatch struct {
	tick *time.Ticker

	OnChange func() error

	// quits loop
	quit    chan bool
	running int32

	Interval time.Duration
}

func NewIntervalReloader(interval time.Duration) *IntervalWatch {
	t := time.NewTicker(interval)
	w := &IntervalWatch{
		tick:     t,
		Interval: interval,
	}
	return w
}

func (w *IntervalWatch) Start() error {
	running := atomic.LoadInt32(&w.running)
	if running == 1 {
		return fmt.Errorf("already running")
	}
	atomic.StoreInt32(&w.running, 1)
	go loop(w)
	return nil

}

func (w *IntervalWatch) Stop() error {
	running := atomic.LoadInt32(&w.running)
	if running == 0 {
		return fmt.Errorf("already stopped")
	}
	atomic.StoreInt32(&w.running, 0)
	w.tick.Stop()
	return nil
}

func (w *IntervalWatch) Stopped() bool {
	running := atomic.LoadInt32(&w.running)
	return (running == 0)
}

func loop(watch *IntervalWatch) {
	defer watch.Stop()
Loop:
	for {
		select {
		case <-watch.quit:
			break Loop
		case <-watch.tick.C:
			err := watch.OnChange()
			if err != nil {
				break Loop
			}
		}
	}
}
