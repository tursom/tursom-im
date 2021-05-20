package utils

import (
	"fmt"
	"io"
)

type Out struct {
	writer io.Writer
}

func (o *Out) Println(i interface{}) {
	fmt.Println()
}

func PrintLn() {

}
