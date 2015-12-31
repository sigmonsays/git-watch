package timer

import (
	"time"
)

func NewIntervalReloader(interval time.Duration, quit chan bool) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			log.Infof("timer reloader: Sending quit signal..\n")
			quit <- true
		}
	}
}
