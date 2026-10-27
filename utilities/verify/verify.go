package main

import (
	"os"

	"wilderzone.org/tai-utilities/internal/frontend"
)

func main() {
	os.Exit(frontend.Run(frontend.Utility{
		Command: "verify",
		Title:   "Verify " + frontend.PRODUCT_NAME,
		Running: "Verifying...",
		Success: "All files have been verified.",
		Failure: "Some files couldn't be verified.",
	}))
}
