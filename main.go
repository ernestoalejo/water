package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		return err
	}
	defer f.Close()

	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	lexer := NewLexer(string(contents))

	go func() {
		for item := range lexer.items {
			fmt.Println(item)
		}
	}()

	for lexer.state != nil {
		lexer.state = lexer.state(lexer)
	}

	return nil
}
