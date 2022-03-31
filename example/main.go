package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/RyuaNerin/uotp"
)

func main() {
	// Create an instance of UOTP
	otp, _ := uotp.New(nil)

	// Issue a new account
	err := otp.Issue(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println("Serial Number:", otp.GetSerialNumber())

	// Sync time with the server
	err = otp.SyncTime(context.Background())
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
