package main

import (
	"github.com/astriaorg/seq-faucet/cmd"
)

//go:generate npm run build-web
func main() {
	cmd.Execute()
}
