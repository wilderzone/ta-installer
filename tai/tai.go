package main

import (
	"fmt"
	"os"
)

func main() {
	arguments := GetArguments()
	status := ""

	err := ValidateArguments(&arguments)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	switch arguments.method {
		case "asset":
			status, err = GenerateAssetFile(arguments)

		case "manifest":
			status, err = GenerateManifestFile(arguments)

		case "install":
			status, err = Install(arguments)

		case "update":
			status, err = Update(arguments)

		case "verify":
			status, err = Verify(arguments)
	}

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	} else {
		fmt.Println(status)
	}
}
