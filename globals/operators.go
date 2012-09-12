package globals

import (
	"fmt"
)

type intFunc func(a, b int) int

var (
	intFuncs = map[string]intFunc{
		"plus": func(a, b int) int { return a + b },
		"minus": func(a, b int) int { return a - b },
		"times": func(a, b int) int { return a * b },
		"divide": func(a, b int) int { return a / b },
	}
)

func op(name string, args []interface{}) (interface{}, error) {
	// Check that two arguments are provided at least
	if len(args) < 2 {
		return 0, fmt.Errorf("at least two params are needed for the %s operator", name)
	}

	// Try to sum the integers
	ac, ok := args[0].(int)
	if ok {
		f := intFuncs[name]
		for _, arg := range args[1:] {
			ac = f(ac, arg.(int))
		}
		return ac, nil
	}

	return 0, fmt.Errorf("%s operator can't handle this kind of numbers", name)
}

func Plus(args ...interface{}) (interface{}, error) {
	return op("plus", args)
}

func Minus(args ...interface{}) (interface{}, error) {
	return op("minus", args)
}

func Times(args ...interface{}) (interface{}, error) {
	return op("times", args)
}

func Divide(args ...interface{}) (interface{}, error) {
	return op("divide", args)
}
