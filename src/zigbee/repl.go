package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// REPL read-evaluate-print loop
type REPL struct {
	Gateway *Gateway
}

// NewREPL called by main.go.
// The actual loop is run (also from main.go)
// as a method off the object created and returned by this.
func NewREPL(gateway *Gateway) *REPL {
	return &REPL{
		Gateway: gateway,
	}
}

// Run The actual loop
// set input to come from console
// output zigbee> prompt, read a string line
// identify command, if successful
// process line as a command with arguments, repeat indefinitely
// (loop exit is expected to be by ctrl-c on the input stream)
// any failures within this loop report an error then continue
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

// NewCommand breaks the input line into space-separated fields.
// if successful, those fields then go into a Command/Args structure.
func NewCommand(line string) *Command {
	fields := strings.Fields(line)

	// All commands are expected to have at least one argument.
	if len(fields) == 0 {
		return nil
	}

	// The single-argument case is treated here
	// to avoid a fields[1:1] range error.
	if len(fields) == 1 {
		command := &Command{
			Command: fields[0],
			Args:    []string{},
		}
		return command
	}

	// All other numbers of arguments.
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

// Run method. Look up the command name supplied by the user
// to see if it exists as a CommandFunc. If it is recognised
// then use the returned value as the command procedure to be executed.
func (c *Command) Run(r *REPL) error {
	if fn, ok := commandFuncs[c.Command]; !ok {
		return fmt.Errorf("command not found")
	} else {
		return fn(r, c)
	}
}
