package main

import (
	"fmt"
	"os"
)

func main() {
	arguments := GetArguments()

	err := ValidateArguments(&arguments)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	logger := Logger{machine: arguments.machine, silent: arguments.silent}

	switch arguments.method {
		case "asset":
			err = GenerateAssetFile(&logger, arguments)

		case "manifest":
			err = GenerateManifestFile(&logger, arguments)

		case "check":
			err = Check(&logger, arguments)

		case "install":
			err = Install(&logger, arguments)

		case "update":
			err = Update(&logger, arguments)

		case "verify":
			err = Verify(&logger, arguments)
	}

	if err != nil {
		logger.error(err.Error())
		os.Exit(1)
	}
}
