//go:build wasm
// +build wasm

package main

import (
	"github.com/syumai/workers"
)

func main() {
	r := setupApp()
	workers.Serve(r)
}
