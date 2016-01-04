package inotify

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"

	"golang.org/x/exp/inotify"
)

func NewInotifyWatch(cfg *InotifyWatchConfig) *InotifyWatch {
	w := &InotifyWatch{
		Dir:  cfg.Dir,
		quit: make(chan bool, 0),
	}
	return w

}

type Event struct {
	*inotify.Event
}

type InotifyWatch struct {
	Dir     string
	running int32
	quit    chan bool

	OnChange func(ev *Event) error
}

type InotifyWatchConfig struct {
	Dir string
}

func DefaultInotifyWatchConfig() *InotifyWatchConfig {
	c := &InotifyWatchConfig{
		Dir: ".",
	}
	return c
}

func (me *InotifyWatch) Start() error {
	running := atomic.LoadInt32(&me.running)
	if running == 1 {
		return fmt.Errorf("already running")
	}
	go loop(me, me.Dir, me.quit)
	atomic.StoreInt32(&me.running, 1)
	return nil
}

func (me *InotifyWatch) Stop() error {
	running := atomic.LoadInt32(&me.running)
	if running == 0 {
		return fmt.Errorf("already stopped")
	}
	atomic.StoreInt32(&me.running, 0)
	return nil
}

func (me *InotifyWatch) Stopped() bool {
	running := atomic.LoadInt32(&me.running)
	return (running == 0)
}

func loop(watch *InotifyWatch, dir string, quit chan bool) {

	watcher, err := inotify.NewWatcher()
	if err != nil {
		log.Warnf("inotify error: %s", err)
		return
	}

	flags := inotify.IN_MODIFY

	err = watcher.AddWatch(dir, flags)
	if err != nil {
		log.Warnf("inotify watch error: %s", err)
		return
	}

	log.Infof("inotify watching %s", dir)

	for {
		select {
		case ev := <-watcher.Event:
			log.Debugf("inotify event %v", ev)
			if strings.HasPrefix(filepath.Base(ev.Name), ".") {
				continue
			}

			event := &Event{ev}
			err := watch.OnChange(event)
			if err != nil {
				log.Warnf("Error updating, skipping reload: %s", err)
				continue
			}
			quit <- true
		case err := <-watcher.Error:
			log.Infof("inotify error: %s", err)
		}

	}

}
