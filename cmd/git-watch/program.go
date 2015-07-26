package main

import (
	"os"
	"os/exec"
)

func program(cmdline string, done, quit chan bool) {
	errchan := make(chan error)
	c := []string{
		"sh",
		"-c",
		cmdline,
	}
	path, err := exec.LookPath(c[0])
	if err != nil {
		done <- true
		return
	}

	cmd := exec.Command(path, c[1:]...)
	log.Infof("exec %s %#v\n", path, c)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	go func() {
		err := cmd.Start()
		if err != nil {
			log.Infof("Command failed to start: %s\n", err)
		}

		err = cmd.Wait()
		if err != nil {
			log.Infof("Command returned error: %s\n", err)
		} else {
			log.Infof("Command returned\n")
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
		log.Infof("Kill process error: %s\n", err)
	}
	done <- true
}
