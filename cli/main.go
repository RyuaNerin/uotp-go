package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/RyuaNerin/uotp"
)

func main() {
	var flagIssue bool
	var flagForce bool
	var confPath string
	var autoSync bool

	flag.BoolVar(&flagIssue, "issue", false, "Issue a new account")
	flag.BoolVar(&flagForce, "force", false, "Never prompt")
	flag.StringVar(&confPath, "conf", "~/.config/uotp/config.json", "Path to the configuration file.")
	flag.BoolVar(&autoSync, "autosync", true, "Automatically synchronize time before generating OTP tokens")
	flag.Parse()

	if confPathEnv := os.Getenv("UOTP_CONF"); confPathEnv != "" {
		confPath = confPathEnv
	}
	confPath = solvePath(confPath)

	_, err := os.Stat(confPath)
	exists := !os.IsNotExist(err)

	var otp uotp.UOTP

	if flagIssue || !exists {
		if exists {
			confirm(flagForce, "Account already exists. Do you want to replace it?")
		} else {
			confirm(flagForce, "Account not exists. Do you want to issue one now?")
		}

		otp, _ = uotp.New(nil)
		if autoSync {
			err = otp.SyncTime(context.Background())
			if err != nil {
				panic(err)
			}
		}
		err = otp.Issue(context.Background())
		if err != nil {
			panic(err)
		}

		save(confPath, otp.GetAccount())

		fmt.Println("A new account has been issued.")
		fmt.Println("Please keep your configuration file safe as it is not possible to recover the account if it gets lost.")
		fmt.Println()
		fmt.Println("Serial Number:", otp.GetSerialNumber())
	} else {
		fs, err := os.Open(confPath)
		if err != nil {
			panic(err)
		}
		defer fs.Close()

		var account uotp.Account
		err = json.NewDecoder(fs).Decode(&account)
		if err != nil && err != io.EOF {
			panic(err)
		}

		otp, err = uotp.New(&account)
		if err != nil {
			panic(err)
		}

		fs.Close()

		fmt.Println("Serial Number:", otp.GetSerialNumber())
	}

	if autoSync {
		err = otp.SyncTime(context.Background())
		if err != nil {
			panic(err)
		}

		save(confPath, otp.GetAccount())
	}

	fmt.Println("OTP Token:", otp.GenerateToken())
}

func confirm(force bool, body string) {
	if force {
		return
	}

	fmt.Println(body, " Y/N")

	b := make([]byte, 1)
	os.Stdin.Read(b)
	if b[0] != 'Y' {
		os.Exit(1)
	}
}

func solvePath(path string) string {
	// https://stackoverflow.com/a/17617721
	usr, _ := user.Current()
	dir := usr.HomeDir

	if path == "~" {
		// In case of "~", which won't be caught by the "else if"
		return dir
	} else if strings.HasPrefix(path, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		return filepath.Join(dir, path[2:])
	}

	return path
}

func save(path string, account uotp.Account) {
	err := os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		panic(err)
	}

	fs, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	defer fs.Close()

	e := json.NewEncoder(fs)
	e.SetIndent("", "  ")
	err = e.Encode(&account)
	if err != nil {
		panic(err)
	}
}
