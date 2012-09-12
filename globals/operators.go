package globals

import (
	"fmt"
)

func Plus(args ...interface{}) (interface{}, error) {
	// Check that two arguments are provided at least
	if len(args) < 2 {
		return 0, fmt.Errorf("at least two params are needed for the plus operator")
	}

	// Try to sum the integers
	_, ok := args[0].(int)
	if ok {
		var ac int
		for _, arg := range args {
			ac += arg.(int)
		}
		return ac, nil
	}

	return 0, fmt.Errorf("num not recognized")
}
