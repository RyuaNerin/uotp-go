This repository is ported of [`devunt/uotp`](https://github.com/devunt/uotp) in Go

# μOTP+

μOTP+ is the next generation OTP toolkit.

## Installation

TODO

## How to use μOTP+ as CLI application

```sh
> uotp
```

## Configuration file

By default, a new configuration file will be automatically generated on ~/.config/uotp/config.json.

This behaviour however can be overriden by passing --conf=/path/to/config.json to uotp command or setting UOTP_CONF=/path/to/config.json environment variable.

```shell
> uotp --conf=uotp.json
> UOTP_CONF=uotp.json uotp
```

## How to develop an application using μOTP+

```go
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
```

## License

All proprietary materials are intellectual property of (C) 2004 - 2017 ATsolutions
