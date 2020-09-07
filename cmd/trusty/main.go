package main

import (
	"fmt"
	"os"
)

const (
	rcError   = 1
	rcSuccess = 0
)

func main() {
	rc := rcSuccess
	fmt.Println("under construction")
	os.Exit(rc)
}
