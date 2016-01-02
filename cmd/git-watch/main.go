package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sigmonsays/git-watch/reload/git"
	"github.com/sigmonsays/go-logging"
)

func Printf(s string, args ...interface{}) {
	log.Infof(s, args...)
}

func main() {

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT)

	done := make(chan bool)
	quit := make(chan bool)
	result := make(chan int)

	var check_interval int
	var configfile = "git-watch.yaml"
	cfg := DefaultGitWatchConfig()

	flag.IntVar(&check_interval, "check", cfg.CheckInterval, "git check interval (seconds)")
	flag.StringVar(&configfile, "config", configfile, "git watch config")
	flag.StringVar(&cfg.Dir, "dir", cfg.Dir, "change directory before starting")
	flag.StringVar(&cfg.LocalBranch, "branch", cfg.LocalBranch, "local branch")
	flag.StringVar(&cfg.ExecCmd, "exec-cmd", cfg.ExecCmd, "exec command")
	flag.StringVar(&cfg.UpdateCmd, "update-cmd", cfg.UpdateCmd, "update command")
	flag.StringVar(&cfg.InstallCmd, "install-cmd", cfg.InstallCmd, "install command")
	flag.StringVar(&cfg.HttpServerAddr, "http", cfg.HttpServerAddr, "start a http server")
	// TODO: env
	flag.BoolVar(&cfg.InheritEnv, "inherit-env", cfg.InheritEnv, "inherit environment")
	flag.StringVar(&cfg.StaticDir, "static-dir", cfg.StaticDir, "static directory")
	flag.StringVar(&cfg.StaticDir, "inotify-dir", cfg.InotifyDir, "use inotify as a trigger in directory")
	flag.StringVar(&cfg.LogLevel, "loglevel", cfg.LogLevel, "set log level")
	flag.BoolVar(&cfg.Once, "once", cfg.Once, "run once and exit")

	flag.Parse()

	if check_interval > 0 {
		cfg.CheckInterval = check_interval
	}

	err := cfg.FromFile(configfile)
	if err != nil {
		log.Infof("error: %s: %s\n", configfile, err)
		return
	}

	cfg.PrintConfig()

	logging.SetLogLevel(cfg.LogLevel)
	logging.SetLogLevels(cfg.LogLevels)

	if cfg.Dir != "" {
		log.Infof("chdir %s\n", cfg.Dir)
		os.Chdir(cfg.Dir)
	}

	if cfg.ExecCmd != "" {
		go program(cfg, cfg.ExecCmd, done, quit)
	}

	var numWatches = 0

	// unique id received map
	watches := make(map[int]*git.GitWatch, 0)

	StartGitWatch := func(id int, cfg *GitWatchConfig) error {
		gw := git.NewGitWatch(cfg.Dir, cfg.LocalBranch)
		gw.Interval = cfg.CheckInterval
		gw.OnChange = func(dir, branch, lhash, rhash string) error {
			err := do_update(cfg)
			if err != nil {
				return err
			}
			quit <- true
			return nil
		}
		gw.OnCheck = func(dir, branch, lhash, rhash string) error {
			result <- id
			return nil
		}
		err = gw.Start()
		if err != nil {
			log.Infof("start: %s", err)
			return err
		}
		numWatches++
		if _, found := watches[id]; found {
			panic("watch already exists")
		}
		watches[id] = gw
		return nil
	}

	if cfg.Dir != "" {
		StartGitWatch(0, cfg)
	}

	// startup other repos
	for n, gitrepo := range cfg.GitRepos {
		var id = n + 1
		cfg_copy := *cfg
		cfg2 := &cfg_copy
		cfg2.Dir = gitrepo

		err = StartGitWatch(id, cfg2)
		if err != nil {
			log.Warnf("start %s: %s", cfg2.Dir, err)
		}

	}

	log.Infof("started %d watched", numWatches)

	if cfg.HttpServerAddr != "" {
		Printf("Starting http server at %s\n", cfg.HttpServerAddr)
		http.Handle("/", http.FileServer(http.Dir(cfg.StaticDir)))

		go func() {
			err := http.ListenAndServe(cfg.HttpServerAddr, nil)
			if err != nil {
				Printf("Starting http server at %s error %s\n", cfg.HttpServerAddr, err)
				os.Exit(1)
			}
		}()
	}

	var remaining = numWatches

Loop:
	for {
		select {
		case id := <-result:
			if cfg.Once {
				watch, found := watches[id]
				if found && watch.Stopped() == false {
					watch.Stop()

					remaining--
					log.Debugf("Received update from watchId:%d remaining:%d", id, remaining)
				}
				if remaining == 0 {
					break Loop
				}
			}

		case <-done:
			log.Infof("Restarting process..\n")
			go program(cfg, cfg.ExecCmd, done, quit)
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
