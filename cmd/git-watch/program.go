package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
)

// parse a command from a string into an array
// if the command starts with [ then its json and we will parse it as an array
//
// by default we treat the string as space delimited and reserve the json syntax for complex
// command lines.
func parse_command(cmdline string) ([]string, error) {
	var c []string
	if strings.HasPrefix(cmdline, "[") {
		err := json.Unmarshal([]byte(cmdline), &c)
		if err != nil {
			return nil, err
		}
	} else {
		c = strings.Split(cmdline, " ")
	}
	return c, nil
}

// used to eat messages when no program is executed (basically a nop)
func noprogram(cfg *GitWatchConfig, cmdline string, done, quit chan bool) {
	for {
		select {
		case <-quit:
		}
	}
}

func program(cfg *GitWatchConfig, cmdline string, done, quit chan bool) {
	errchan := make(chan error)

	if len(cfg.Env) == 0 {
		cfg.Env = os.Environ()
	}

	var c []string
	Printf("cmdline %v\n", cmdline)

	c, err := parse_command(cmdline)
	if err != nil {
		Printf("parsing command: %s: %s\n", cmdline, err)
		done <- false
		return
	}

	path, err := exec.LookPath(c[0])
	if err != nil {
		Printf("LookupPath: %s: %s\n", c[0], err)
		done <- false
		return
	}

	cmd := exec.Command(path, c[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if cfg.InheritEnv == false {
		log.Tracef("Setting custom environment items:%d\n", len(cfg.Env))
		cmd.Env = cfg.Env
	}

	go func() {
		err := cmd.Start()
		if err != nil {
			Printf("Command failed to start: %s\n", err)
		}

		err = cmd.Wait()
		if err != nil {
			Printf("Command returned error: %s\n", err)
		} else {
			Printf("Command returned\n")
		}
		errchan <- err
	}()

Loop:
	for {
		select {
		case <-errchan:
			break Loop

		case <-quit:
			break Loop
		}
	}

	err = cmd.Process.Kill()
	if err != nil {
		Printf("Kill process error: %s\n", err)
	}
	done <- true
}
