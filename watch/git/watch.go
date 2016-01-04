package git

import (
	"fmt"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"
)

type ChangeFunc func(dir, branch, lhash, rhash string) error
type CheckFunc func(dir, branch, lhash, rhash string) error

type GitWatch struct {
	Dir      string
	Branch   string
	Interval int

	running int32
	quit    chan bool

	OnChange ChangeFunc
	OnCheck  CheckFunc
}

func NewGitWatch(dir, branch string) *GitWatch {
	w := &GitWatch{
		Dir:    dir,
		Branch: branch,
		quit:   make(chan bool, 0),
	}
	return w
}

func (w *GitWatch) Start() error {
	err := make(chan error, 1)
	go loop(w, err)
	res := <-err
	if res == nil {
		atomic.StoreInt32(&w.running, 1)
	}
	return res
}
func (w *GitWatch) Stop() error {
	running := atomic.LoadInt32(&w.running)
	if running == 0 {
		return fmt.Errorf("already stopped")
	}
	atomic.StoreInt32(&w.running, 0)
	w.quit <- true
	return nil
}

func (w *GitWatch) Stopped() bool {
	running := atomic.LoadInt32(&w.running)
	return (running == 0)
}

// get a commit hash for branch in git repo at directory
func local_hash(dir, branch string) string {
	cmdline := []string{
		"-C", dir,
		"rev-list",
		"--max-count=1",
		branch,
	}
	out, err := exec.Command("git", cmdline...).Output()
	if err != nil {
		log.Infof("remote_hash: [cmdline %s] error: %s\n", cmdline, err)
		return ""
	}
	return strings.Trim(string(out), "\n")
}

func remote_hash(dir, branch string) string {
	cmdline := []string{
		"-C", dir,
		"ls-remote",
		"origin",
		"-h",
		fmt.Sprintf("refs/heads/%s", branch),
	}
	out, err := exec.Command("git", cmdline...).Output()
	if err != nil {
		log.Infof("remote_hash: [cmdline %s] error: %s\n", cmdline, err)
		return ""
	}
	tmp := strings.Fields(string(out))
	if len(tmp) < 1 {
		return ""
	}
	return tmp[0]
}

func loop(watch *GitWatch, err_chan chan error) {
	dir := watch.Dir
	branch := watch.Branch
	interval := watch.Interval

	if dir == "" {
		dir = "."
	}

	err_chan <- nil

	log.Infof("watching %s directory for changes", dir)

	reload_interval := time.Duration(interval) * time.Second
	log.Debugf("dir:%s reload interval %s", dir, reload_interval)
	reload := time.NewTicker(reload_interval)
Loop:
	for {
		select {
		case <-watch.quit:
			break Loop
		case <-reload.C:
			log.Tracef("Checking for changes in git path %s\n", dir)
			rhash := remote_hash(dir, branch)
			lhash := local_hash(dir, branch)

			log.Debugf("dir: %s: hash lhash:%s rhash:%s", dir, lhash, rhash)

			err := watch.OnCheck(dir, branch, lhash, rhash)
			if err != nil {
				log.Infof("dir=%s: OnCheck: %s\n", dir, err)
				continue
			}

			if len(rhash) == 0 || len(lhash) == 0 {
				continue
			}

			if rhash != lhash {
				log.Infof("dir=%s: Code change detected, remote hash %s != local %s\n", dir, rhash, lhash)
				err = watch.OnChange(dir, branch, lhash, rhash)
				if err != nil {
					log.Infof("dir=%s: Error updating code, skipping reload: %s\n", dir, err)
					continue
				}
			}
		}
	}
}
