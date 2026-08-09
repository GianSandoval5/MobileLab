package main

import (
	"context"
	"os"

	"github.com/mobilelab-dev/mobilelab/internal/cli"
)

func main() {
	os.Exit(cli.Main(context.Background(), os.Args[1:]))
}
