package main

import (
	"fmt"
	"os"
)

func main() {
	arguments, err := getArguments()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		usage(os.Stderr)
		os.Exit(1)
	}

	err = validateArguments(&arguments)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	logger := Logger{machine: arguments.machine, silent: arguments.silent, verbose: arguments.verbose}

	switch arguments.method {
		case "asset":
			err = GenerateAssetFile(&logger, arguments)

		case "manifest":
			err = GenerateManifestFile(&logger, arguments)

		case "check":
			err = Check(&logger, arguments)

		case "check-repo":
			err = CheckRepo(&logger, arguments)

		case "check-updates":
			err = CheckUpdates(&logger, arguments)

		case "install":
			err = Install(&logger, arguments)

		case "list":
			err = List(&logger, arguments)

		case "uninstall":
			err = Uninstall(&logger, arguments)

		case "update":
			err = Update(&logger, arguments)

		case "verify":
			err = Verify(&logger, arguments)
	}

	if err != nil {
		logger.fail(err)
		os.Exit(1)
	}
}
