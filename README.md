# swish

Swish will modify your SSH config file to work with a different identity file
and user.

I use it to switch between SSH accounts for Github.com. I have
a `personal-github` script set up to run this:

```bash
${GOPATH}/bin/swish --identity-file ${HOME}/.ssh/github_rsa --user kevinburke
```

And a `work-github` that switches to my work Github account and profile.

Swish [uses `ssh_config`, a Go SSH config file parser][ssh_config].

[ssh_config]: https://github.com/kevinburke/ssh_config

## Installation

You need a working Go installation, then run:

```bash
go get github.com/kevinburke/swish
```

## Usage

```
Usage of swish:
  -host string
    	Host (default "github.com")
  -identity-file string
    	Identity file
  -user string
    	SSH User
```
