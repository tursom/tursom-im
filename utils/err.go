package utils

import (
	"fmt"
	"io"
	"os"
	"runtime"
)

type Out struct {
	writer io.Writer
}

func (o *Out) Println(i interface{}) {
	fmt.Println()
}

func PrintLn() {

}

func Recover(panicCaused func()) {
	if err := recover(); err != nil {
		if panicCaused != nil {
			panicCaused()
		}
		fmt.Fprintln(os.Stderr, err)
		for i := 1; ; i++ {
			pc, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			fmt.Fprintln(os.Stderr, pc, file, line)
		}
	}
}
