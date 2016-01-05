package main

import (
	"bytes"
	"fmt"
	"launchpad.net/goyaml"
	"os"
)

type RepoSpec struct {
	// directory where the repo should live
	Directory string

	// the remote we'll clone from
	Origin string `yaml:"origin,omitempty"`
}

type GitWatchConfig struct {
	LogLevel  string
	LogLevels map[string]string

	CheckInterval int

	Dir string
	Env []string

	OriginTemplates map[string]string `yaml:"origin_templates"`
	Repositories    []*RepoSpec

	GitRepos []string

	InheritEnv bool

	ExecCmd     string
	UpdateCmd   string
	InstallCmd  string
	LocalBranch string

	HttpServerAddr string
	StaticDir      string
	InotifyDir     string
	Once           bool
	IntervalReload int
}

func DefaultGitWatchConfig() *GitWatchConfig {
	return &GitWatchConfig{
		LogLevel: "error",
		// LogLevels:
		CheckInterval: 5,
		Dir:           ".",
		// GitRepos:
		// InheritEnv:
		ExecCmd:        "make run",
		UpdateCmd:      "make",
		InstallCmd:     "make install",
		LocalBranch:    "master",
		HttpServerAddr: "",
		StaticDir:      "static",
		InotifyDir:     "",
		GitRepos:       make([]string, 0),
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
