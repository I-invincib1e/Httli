package main

import (
	"os"

	"github.com/I-invincib1e/http-cli/cmd"
)

func main() {
	cmd.Execute(os.Args[1:])
}
