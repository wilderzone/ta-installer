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
		Success: frontend.PRODUCT_NAME + " has been verified.",
		Failure: "Couldn't verify " + frontend.PRODUCT_NAME + ".",
	}))
}
