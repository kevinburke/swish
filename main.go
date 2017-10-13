package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"

	"github.com/dchest/safefile"
	"github.com/kevinburke/ssh_config"
)

func setHost(host string, identityFile string, username string) (error, string) {
	user, err := user.Current()
	if err != nil {
		return err, "finding current user"
	}
	f, err := os.Open(filepath.Join(user.HomeDir, ".ssh", "config"))
	if err != nil {
		return err, "opening ssh config file"
	}
	defer f.Close()
	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return err, "parsing ssh config file"
	}
	for _, host := range cfg.Hosts {
		if !host.Matches("github.com") {
			continue
		}
		for _, node := range host.Nodes {
			switch tnode := node.(type) {
			case *ssh_config.KV:
				if tnode.Key == "IdentityFile" {
					tnode.Value = identityFile
				}
				if tnode.Key == "User" {
					tnode.Value = username
				}
			}
		}
	}
	stat, err := f.Stat()
	if err != nil {
		return err, "getting file mode"
	}
	f2, err := safefile.Create(f.Name(), stat.Mode())
	if err != nil {
		return err, "creating temp file"
	}
	data, err := cfg.MarshalText()
	if err != nil {
		return err, "marshaling ssh config"
	}
	_, writeErr := f2.Write(data)
	if writeErr != nil {
		return writeErr, "writing data to temp file"
	}
	commitErr := f2.Commit()
	if commitErr != nil {
		return commitErr, "committing file"
	}
	return nil, ""
}

func checkError(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %s: %v\n", msg, err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "usage: swish child\n")
		os.Exit(1)
	}
	err, msg := setHost("github.com", "~/.ssh/otto-github", "kevinburkeotto")
	checkError(err, msg)
	var args []string
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}
	cmd := exec.Command(os.Args[1], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	c := make(chan os.Signal, 100)
	signal.Notify(c)
	if err := cmd.Start(); err != nil {
		hostErr, msg := setHost("github.com", "~/.ssh/github_rsa", "kevinburke")
		checkError(hostErr, msg)

		fmt.Fprintf(os.Stderr, "Command %q exited on start: %v\n", os.Args[1], err)
		os.Exit(1)
	}
	done := make(chan error, 1)
	go func(ch chan<- error) {
		ch <- cmd.Wait()
	}(done)
	for {
		select {
		case sig := <-c:
			if err := cmd.Process.Signal(sig); err != nil {
				fmt.Fprintf(os.Stderr, "could not send signal %q to subprocess: %v\n", err)
			}
			continue
		case err := <-done:
			hostErr, msg := setHost("github.com", "~/.ssh/github_rsa", "kevinburke")
			checkError(hostErr, msg)
			if err != nil {
				if exiterr, ok := err.(*exec.ExitError); ok {
					// The program has exited with an exit code != 0

					// This works on both Unix and Windows. Although package
					// syscall is generally platform dependent, WaitStatus is
					// defined for both Unix and Windows and in both cases has
					// an ExitStatus() method with the same signature.
					if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
						os.Exit(status.ExitStatus())
					}
				}
				fmt.Fprintf(os.Stderr, "Command %q exited with unknown error: %v\n", os.Args[1], err)
				os.Exit(1)
			}
			return
		}
	}
}
