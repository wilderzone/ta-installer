package main

import (
	"os"

	"wilderzone.org/tai-utilities/internal/frontend"
)

func main() {
	os.Exit(frontend.Run(frontend.Utility{
		Command:     "update",
		Title:       "Update " + frontend.PRODUCT_NAME,
		Running:     "Checking for updates...",
		Success:     frontend.PRODUCT_NAME + " has been updated.",
		Failure:     "Couldn't update " + frontend.PRODUCT_NAME + ".",
		Conclusions: map[string]string{
			"INSTALL_UP_TO_DATE": frontend.PRODUCT_NAME + " is already up to date.",
		},
	}))
}
