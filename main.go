package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	if err := repl(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}

func repl() error {
	backend := NewBackend()
	inputReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input, err := inputReader.ReadString('\n')
		if err != nil {
			return err
		}
		err = execute(input, backend)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}
}

func execute(input string, backend *Backend) error {
	lexer := NewLexer()
	parser := NewParser()

	tokens := lexer.Scan(input)
	statement, err := parser.Parse(tokens)
	if err != nil {
		return err
	}

	backend.Run(statement)

	return nil
}
