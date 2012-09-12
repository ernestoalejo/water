package globals

import (
	"fmt"
)

func Print(format string, args ...interface{}) string {
	return fmt.Sprintf(format + "\n", args...)
}
