package main

import (
	"os"

	"wilderzone.org/tai-utilities/internal/frontend"
)

func main() {
	os.Exit(frontend.Run(frontend.UPDATE))
}
