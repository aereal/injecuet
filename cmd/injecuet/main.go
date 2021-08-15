package main

import (
	"os"

	"github.com/aereal/injecuet/internal/cli"
)

func main() {
	app := &cli.App{}
	os.Exit(app.Run(os.Args))
}
