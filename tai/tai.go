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
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	logger := Logger{machine: arguments.machine}

	switch arguments.method {
		case "asset":
			status, err = GenerateAssetFile(&logger, arguments)

		case "manifest":
			status, err = GenerateManifestFile(&logger, arguments)

		case "install":
			status, err = Install(&logger, arguments)

		case "update":
			status, err = Update(&logger, arguments)

		case "verify":
			status, err = Verify(&logger, arguments)
	}

	if err != nil {
		logger.error(err.Error())
		os.Exit(1)
	} else {
		logger.print(status)
	}
}
