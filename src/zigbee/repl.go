package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type REPL struct {
	Gateway *Gateway
}

func NewREPL(gateway *Gateway) *REPL {
	return &REPL{
		Gateway: gateway,
	}
}

func (r *REPL) Run() {
	stdin := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("zigbee> ")

		if line, err := stdin.ReadString('\n'); err != nil {
			fmt.Println(err)
			return
		} else if command := NewCommand(line); command == nil {

		} else if err := command.Run(r); err != nil {
			fmt.Println(err)
		}
	}
}

type Command struct {
	Command string
	Args    []string
}

func NewCommand(line string) *Command {
	fields := strings.Fields(line)

	if len(fields) == 0 {
		return nil
	}

	if len(fields) == 1 {
		command := &Command{
			Command: fields[0],
			Args:    []string{},
		}
		return command
	}

	command := &Command{
		Command: fields[0],
		Args:    fields[1:],
	}
	return command
}

type CommandFunc func(r *REPL, c *Command) error

var commandFuncs = map[string]CommandFunc{
	"add":    AddCommand,
	"remove": RemoveCommand,
	"list":   ListCommand,
}

func (c *Command) Run(r *REPL) error {
	if fn, ok := commandFuncs[c.Command]; !ok {
		return fmt.Errorf("command not found")
	} else {
		return fn(r, c)
	}
}
