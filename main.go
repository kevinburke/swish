package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/dchest/safefile"
	"github.com/kevinburke/ssh_config"
)

func setHost(configFile, host, identityFile, username string) (error, string) {
	f, err := os.Open(configFile)
	if err != nil {
		return err, "opening ssh config file"
	}
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
	name := f.Name()
	if err := f.Close(); err != nil {
		return err, "closing file"
	}
	f2, err := safefile.Create(name, stat.Mode())
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
	identityFile := flag.String("identity-file", "", "Identity file")
	userArg := flag.String("user", "", "SSH User")
	host := flag.String("host", "github.com", "Host")
	flag.Parse()
	if *userArg == "" && *identityFile == "" {
		checkError(errors.New("please provide a user or identity file"), "parsing command line arguments")
	}
	u, err := user.Current()
	if err != nil {
		checkError(err, "finding current user")
	}
	configFile := filepath.Join(u.HomeDir, ".ssh", "config")
	err, msg := setHost(configFile, *host, *identityFile, *userArg)
	checkError(err, msg)
}
