package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	if err := runCLI(context.Background(), os.Args[1:], os.Getenv, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
