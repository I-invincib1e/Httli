package main

import (
	"os"

	"github.com/I-invincib1e/httli/cmd"
)

func main() {
	cmd.Execute(os.Args[1:])
}
