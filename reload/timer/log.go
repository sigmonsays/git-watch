package timer

import (
	gologging "github.com/sigmonsays/go-logging"
)

var log gologging.Logger

func init() {
	log = gologging.Register("timer", func(newlog gologging.Logger) { log = newlog })
}
