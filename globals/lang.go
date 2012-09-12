package globals

import (
)

func Begin(args ...interface{}) interface{} {
	if len(args) == 0 {
		return nil
	}

	return args[len(args)-1]
}
