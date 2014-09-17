package main
import (
    "fmt"
    "time"
    "strings"
    "os"
    "os/exec"
)
func local_hash(cfg *GitWatchConfig) string {
    revlist := fmt.Sprintf("rev-list --max-count=1 %s", cfg.LocalBranch)
    cmdline := strings.Split(revlist, " ")
    out, err := exec.Command("git", cmdline...).Output()
    if err != nil {
        Printf("remote_hash: ls-remote error: %s\n", err)
        return ""
    }
    return strings.Trim(string(out), "\n")
}
func remote_hash(cfg *GitWatchConfig) string {
    lsremote := "ls-remote origin -h refs/heads/master"
    cmdline := strings.Split(lsremote, " ")
    out, err := exec.Command("git", cmdline...).Output()
    if err != nil {
        Printf("remote_hash: ls-remote error: %s\n", err)
        return ""
    }
    tmp := strings.Fields(string(out))
    return tmp[0]
}

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
    err := run_command("git", "pull")
    if err != nil {
        return err
    }

    err = run_command(cfg.UpdateCmd)
    if err != nil {
        return err
    }
    err = run_command(cfg.InstallCmd)
    if err != nil {
        return err
    }
    return err
}

func git_watch(cfg *GitWatchConfig, quit chan bool) {
    for {
        select {
        case <- time.After(time.Duration(cfg.CheckInterval) * time.Second):
            Printf("Checking for changes in git...\n")
            rhash := remote_hash(cfg)
            lhash := local_hash(cfg)

            if len(rhash) == 0 || len(lhash) == 0 {
                continue
            }

            if rhash != lhash {
                Printf("Code change detected, remote hash %s != local %s\n", rhash, lhash)
                err := do_update(cfg)
                if err != nil {
                    Printf("Error updating code, skipping reload: %s\n", err)
                    continue
                }
                quit <- true
            }
        }
    }
    
}
