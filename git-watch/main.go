package main
import (
    "fmt"
    "flag"
    "syscall"
    "os"
    "os/signal"
)

func Printf(s string, args...interface{}) {
    fmt.Printf("[git-watch] " + s, args...)
}

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
      Printf("error: %s: %s\n", configfile, err)
      return
    }

    cfg.PrintConfig()

    if cfg.Dir != "" {
      Printf("chdir %s\n", cfg.Dir)
      os.Chdir(cfg.Dir)
    }

    go program(cfg.ExecCmd, done, quit)

    go git_watch(cfg, quit)

    Loop:
    for {
        select {
        case <- done:
            Printf("Restarting process..\n")
            go program(cfg.ExecCmd, done, quit)
        case signum := <- sig:
            Printf("Got signal %d\n", signum)
            if signum == syscall.SIGHUP {
                quit <- true
            } else if signum == syscall.SIGINT {
                break Loop
            } else {
                Printf("Unhandled signal received: %d\n", signum)
            }
        }
    }
    
}
