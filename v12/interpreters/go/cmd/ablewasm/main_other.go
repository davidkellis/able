//go:build !js || !wasm

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "ablewasm is only available with GOOS=js GOARCH=wasm")
	os.Exit(1)
}
