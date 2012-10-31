package globals

import (
	"fmt"
)

func Print(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func Println(args ...interface{}) {
	fmt.Println(args...)
}
