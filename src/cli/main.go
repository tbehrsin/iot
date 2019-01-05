package main

import (
	"fmt"
	"os"
)

type CreateTokenRequest struct {
	Code string `json:"code"`
}

type CreateTokenResponse struct {
	Token string `json:"token"`
}

func Main(command string, args ...string) {
	var config *Config
	var err error
	if config, err = LoadConfig(); err != nil {
		panic(err)
	}
	defer config.Save()

	switch command {
	case "login":
		LoginCommand(config, args)
	case "serve":
		ServeCommand(config, args)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Print(`
Available commands:

  create
  login
  init
  serve

`)
	} else {
		Main(os.Args[1], os.Args[2:]...)
	}
}
