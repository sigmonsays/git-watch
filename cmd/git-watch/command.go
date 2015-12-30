package main

import (
	"os"
	"os/exec"
	"strings"
)

func run_command(name string, cmdline ...string) error {
	Printf("[command] %s %s\n", name, strings.Join(cmdline, " "))
	cmd := exec.Command(name, cmdline...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func do_update(cfg *GitWatchConfig) error {
	err := run_command("git", "-C", cfg.Dir, "pull")
	if err != nil {
		return err
	}

	if len(cfg.UpdateCmd) > 0 {
		update_cmd, err := parse_command(cfg.UpdateCmd)
		if err != nil {
			return err
		}
		Printf("UpdateCmd %v\n", update_cmd)
		err = run_command(update_cmd[0], update_cmd[1:]...)
		if err != nil {
			return err
		}
	}

	if len(cfg.InstallCmd) > 0 {
		install_cmd, err := parse_command(cfg.InstallCmd)
		if err != nil {
			return err
		}
		Printf("InstallCmd %v\n", install_cmd[1:])
		err = run_command(install_cmd[0], install_cmd[1:]...)
		if err != nil {
			return err
		}
	}
	return err
}
