package main

import (
	"flag"
	"fmt"
	"io"
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
	// The source stream
	var f io.ReadCloser

	if flag.Arg(0) != "" {
		// Open the file if it's the first arg
		var err error
		f, err = os.Open(flag.Arg(0))
		if err != nil {
			return err
		}
		defer f.Close()
	} else {
		// Use stdin if there's no filename in the args
		f = os.Stdin
	}

	// Parse it
	root, err := Parse(f)
	if err != nil {
		return err
	}

	// Start the global funcs
	funcs := initGlobalFuncs()

	// Exec it
	if err := Exec(os.Stdout, root, funcs); err != nil {
		return err
	}

	return nil
}

func initGlobalFuncs() map[string]interface{} {
	return map[string]interface{}{
		"+":       globals.Plus,
		"-":       globals.Minus,
		"*":       globals.Times,
		"/":       globals.Divide,
		"%":       globals.Modulo,
		">":       globals.GreaterThan,
		">=":      globals.GreaterEqual,
		"<":       globals.LessThan,
		"<=":      globals.LessEqual,
		"=":       globals.Equal,
		"print":   globals.Print,
		"println": globals.Println,
		"not":     globals.Not,
	}
}
