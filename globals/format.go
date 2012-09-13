package globals

import (
	"fmt"
)

func Print(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func Println(args ...interface{}) string {
	return fmt.Sprintln(args...)
}
