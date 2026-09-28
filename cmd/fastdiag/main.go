package main

import (
	"fmt"
	"os"
)

func main() {
	if err := execute(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "fastdiag 執行失敗：%v\n", err)
		os.Exit(1)
	}
}
