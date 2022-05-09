package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/RyuaNerin/uotp"
)

func main() {
	var otp uotp.UOTP

	if fs, err := os.Open("uotp.json"); !os.IsNotExist(err) {
		defer fs.Close()

		// Read
		var otpAccount uotp.Account
		err = json.NewDecoder(fs).Decode(&otpAccount)
		if err != nil {
			panic(err)
		}

		otp, err = uotp.New(&otpAccount)
		if err != nil {
			panic(err)
		}
	} else {
		// Create an instance of UOTP
		otp, _ = uotp.New(nil)

		// Issue a new account
		err := otp.Issue(context.Background())
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Serial Number:", otp.GetSerialNumber())

	// Sync time with the server
	err := otp.SyncTime(context.Background())
	if err != nil {
		panic(err)
	}

	// Get a new OTP token
	fmt.Println("OTP Token:", otp.GenerateToken())

	// Save UOTP Acocunt
	fs, err := os.Create("uotp.json")
	if err != nil {
		panic(err)
	}
	defer fs.Close()

	err = json.NewEncoder(fs).Encode(otp.GetAccount())
	if err != nil {
		panic(err)
	}
}
