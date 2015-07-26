package main

import (
	"time"
)

func timer_reloader(quit chan bool) {
	for {
		select {
		case <-time.After(5 * time.Second):
			log.Infof("timer reloader: Sending quit signal..\n")
			quit <- true
		}
	}

}
