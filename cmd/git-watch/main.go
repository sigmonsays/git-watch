package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sigmonsays/git-watch/reload/git"
)

func main() {

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT)

	done := make(chan bool)
	quit := make(chan bool)

	var check_interval int
	var configfile string
	flag.IntVar(&check_interval, "check", 30, "git check interval")
	flag.StringVar(&configfile, "config", "git-watch.yaml", "git watch config")
	flag.Parse()

	cfg := DefaultGitWatchConfig()
	cfg.CheckInterval = check_interval

	err := cfg.FromFile(configfile)
	if err != nil {
		log.Infof("error: %s: %s\n", configfile, err)
		return
	}

	cfg.PrintConfig()

	if cfg.Dir != "" {
		log.Infof("chdir %s\n", cfg.Dir)
		os.Chdir(cfg.Dir)
	}

	go program(cfg.ExecCmd, done, quit)

	gw := git.NewGitWatch(cfg.Dir, cfg.LocalBranch)
	gw.OnChange = func(dir, branch, lhash, rhash string) error {
		quit <- true
		return nil
	}

	err = gw.Start()
	if err != nil {
		log.Infof("start: %s", err)
		return
	}

Loop:
	for {
		select {
		case <-done:
			log.Infof("Restarting process..\n")
			go program(cfg.ExecCmd, done, quit)
		case signum := <-sig:
			log.Infof("Got signal %d\n", signum)
			if signum == syscall.SIGHUP {
				quit <- true
			} else if signum == syscall.SIGINT {
				break Loop
			} else {
				log.Infof("Unhandled signal received: %d\n", signum)
			}
		}
	}

}