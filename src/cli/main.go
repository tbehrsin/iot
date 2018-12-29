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

func main() {
	var config *Config
	var err error
	if config, err = LoadConfig(); err != nil {
		panic(err)
	}
	defer config.Save()

	if len(os.Args) < 2 {
		fmt.Print(`
Available commands:

  create
  login
  init
  serve

`)
	} else {
		switch os.Args[1] {
		case "login":
			LoginCommand(config, os.Args[2:])
		case "serve":
			ServeCommand(config, os.Args[2:])
		}
	}
}
