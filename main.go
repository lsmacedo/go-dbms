package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/goombaio/namegenerator"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}

func run() error {
	var inputStrings []string

	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)

	inputStrings = append(inputStrings, "create table users id integer, name text, age integer")

	for i := 0; i < 10; i++ {
		input := fmt.Sprintf("insert into users id, name, age values %d, '%s', %d", i, nameGenerator.Generate(), 15+rand.Intn(50))
		inputStrings = append(inputStrings, input)
	}

	backend := NewBackend()
	for i := range inputStrings {
		err := execute(inputStrings[i], backend)
		if err != nil {
			return err
		}
	}

	return repl(backend)
}

func repl(backend *Backend) error {
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
