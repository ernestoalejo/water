package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		fmt.Println("ERROR:", err)
	}
}

func run() error {
	// Open the file
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		return err
	}
	defer f.Close()

	// Parse it
	_, err = Parse(f)
	if err != nil {
		return err
	}

	fmt.Println("finished!")

	return nil
}
