package inotify

import (
	gologging "github.com/sigmonsays/go-logging"
)

var log gologging.Logger

func init() {
	log = gologging.Register("inotify", func(newlog gologging.Logger) { log = newlog })
}
