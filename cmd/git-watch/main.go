package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"text/template"
	"time"

	"github.com/sigmonsays/git-watch/watch"
	"github.com/sigmonsays/git-watch/watch/git"
	"github.com/sigmonsays/git-watch/watch/inotify"
	"github.com/sigmonsays/git-watch/watch/timer"
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
	result := make(chan string)

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

	if log.IsDebug() {
		cfg.PrintConfig()
	}

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
	watches := make(map[string]watch.Watcher, 0)

	if cfg.InotifyDir != "" {
		icfg := inotify.DefaultInotifyWatchConfig()
		icfg.Dir = cfg.InotifyDir
		iw := inotify.NewInotifyWatch(icfg)
		err := iw.Start()
		if err != nil {
			log.Infof("start: %s", err)
		}
		watches["inotify"] = iw
		iw.OnChange = func(ev *inotify.Event) error {
			quit <- true
			return nil
		}
	}

	if cfg.IntervalReload > 0 {

		iw := timer.NewIntervalReloader(time.Duration(cfg.IntervalReload) * time.Second)
		err := iw.Start()
		if err != nil {
			log.Infof("start: %s", err)
		}
		iw.OnChange = func() error {
			quit <- true
			return nil
		}
		watches["timer"] = iw

	}

	StartGitWatch := func(id string, cfg *GitWatchConfig) error {
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
		StartGitWatch("dir", cfg)
	}

	// mapping of directory to repo spec
	repo_paths := make(map[string]*RepoSpec, 0)

	// merge in .GitRepos into the repositories struct
	for _, gitrepo := range cfg.GitRepos {
		repo, found := repo_paths[gitrepo]
		if found {

			continue
		}

		repo = &RepoSpec{
			Directory: gitrepo,
			// Origin: left blank intentionally
		}
		cfg.Repositories = append(cfg.Repositories, repo)
	}

	// startup other repos
	for n, repo := range cfg.Repositories {
		var id = n + 1
		name := fmt.Sprintf("watch-%d", id)
		cfg_copy := *cfg
		cfg2 := &cfg_copy
		cfg2.Dir = repo.Directory

		// clone repo if it does not exist
		var doSetup bool
		st, err := os.Stat(repo.Directory)
		if err != nil && os.IsNotExist(err) {
			doSetup = true
		}

		if err == nil && st.IsDir() == false {
			doSetup = false
		}

		if doSetup {
			err = setupRepo(cfg, repo)
			if err != nil {
				continue
			}
		}

		err = StartGitWatch(name, cfg2)
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

type CloneTemplate struct {
	Basename string
}

func setupRepo(cfg *GitWatchConfig, repo *RepoSpec) error {

	var remote string

	if repo.Origin == "" {

		origin_template, found := cfg.OriginTemplates["default"]
		if found == false {
			log.Warnf("unable to setup repo %s: no default origin template to clone from", repo.Directory)
			return nil
		}
		tmpl := template.Must(template.New("template").Parse(origin_template))

		tmplData := &CloneTemplate{
			Basename: filepath.Base(repo.Directory),
		}

		var buf bytes.Buffer
		err := tmpl.Execute(&buf, tmplData)
		if err != nil {
			return nil
		}

		remote = buf.String()

	} else {
		remote = repo.Origin
	}

	cmdline := []string{
		"git",
		"clone",
		remote,
		repo.Directory,
	}
	log.Infof("clone %s", cmdline)

	cmd := exec.Command(cmdline[0], cmdline[1:]...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	return err
}
