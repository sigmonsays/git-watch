package main

import (
	"bytes"
	"fmt"
	"launchpad.net/goyaml"
	"os"
)

type GitWatchConfig struct {
	LogLevel  string
	LogLevels map[string]string

	CheckInterval int

	Dir string
	Env []string

	GitRepos []string

	InheritEnv bool

	ExecCmd     string
	UpdateCmd   string
	InstallCmd  string
	LocalBranch string
}

func DefaultGitWatchConfig() *GitWatchConfig {
	return &GitWatchConfig{
		CheckInterval: 1,
		LocalBranch:   "master",
	}
}

func (cfg *GitWatchConfig) FromFile(path string) error {
	err := cfg.LoadYaml(path)
	return err
}

func (c *GitWatchConfig) LoadYaml(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	b := bytes.NewBuffer(nil)
	_, err = b.ReadFrom(f)
	if err != nil {
		return err
	}

	if err := c.LoadYamlBuffer(b.Bytes()); err != nil {
		return err
	}

	if err := c.FixupConfig(); err != nil {
		return err
	}

	return nil
}
func (c *GitWatchConfig) LoadYamlBuffer(buf []byte) error {
	err := goyaml.Unmarshal(buf, c)
	if err != nil {
		return err
	}
	return nil
}

func (conf *GitWatchConfig) PrintConfig() {
	d, err := goyaml.Marshal(conf)
	if err != nil {
		fmt.Println("Marshal error", err)
		return
	}
	fmt.Println(string(d))
}

func (c *GitWatchConfig) FixupConfig() error {

	return nil
}
