package git

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type ChangeFunc func(dir, branch, lhash, rhash string) error

type GitWatch struct {
	Dir      string
	Branch   string
	Interval int

	OnChange ChangeFunc
}

func NewGitWatch(dir, branch string) *GitWatch {
	w := &GitWatch{
		Dir:    dir,
		Branch: branch,
	}
	return w
}

func (w *GitWatch) Start() error {
	err := make(chan error, 1)
	go loop(w, err)
	return <-err
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
	log.Debugf("reload interval %s", reload_interval)
	reload := time.NewTicker(reload_interval)
	for {
		select {
		case <-reload.C:
			log.Debugf("Checking for changes in git path %s\n", dir)
			rhash := remote_hash(dir, branch)
			lhash := local_hash(dir, branch)

			if len(rhash) == 0 || len(lhash) == 0 {
				continue
			}

			if rhash != lhash {
				log.Infof("Code change detected, remote hash %s != local %s\n", rhash, lhash)
				err := watch.OnChange(dir, branch, lhash, rhash)
				if err != nil {
					log.Infof("Error updating code, skipping reload: %s\n", err)
					continue
				}
			}
		}
	}
}
