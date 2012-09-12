package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ernestokarim/water/globals"
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

	log.Println(" * Parsing the file...")

	// Parse it
	root, err := Parse(f)
	if err != nil {
		return err
	}

	// Start the global funcs
	funcs := initGlobalFuncs()

	log.Println(" * Exec the file...")

	// Exec it
	if err := Exec(os.Stdout, root, funcs); err != nil {
		return err
	}

	return nil
}

func initGlobalFuncs() map[string]interface{} {
	return map[string]interface{}{
		"+": globals.Plus,
	}
}
