package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func LoginCommand(config *Config, args []string) {
	var response CreateTokenResponse
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter code: ")
	code, _ := reader.ReadString('\n')
	code = strings.TrimSpace(code)

	gateway := FindGateway("")

	if err := gateway.Post("/api/v1/auth/", &CreateTokenRequest{code}, &response); err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		config.Gateway = gateway.ID
		config.Token = response.Token
	}
}

func ServeCommand(config *Config, args []string) {
	fmt.Print("Connecting to gateway...")
	for {
		gateway := FindGateway(config.Gateway)
		if gateway == nil {
			continue
		}
		gateway.Token = config.Token
		gateway.Port = 443

		if conn, err := gateway.WebSocket(); err != nil {
			continue
		} else {
			fmt.Println(" connected.")
			dir, _ := os.Getwd()
			Serve(dir, conn)
			fmt.Println("Disconnected.")
			time.Sleep(1 * time.Second)
			fmt.Print("Reconnecting to gateway...")
		}
	}
}
