package main

import (
	"context"
	"errors"
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

	logger := Logger{machine: arguments.machine, progress: arguments.progress, silent: arguments.silent, verbose: arguments.verbose}

	// Commands that can run for a long time take a context, so they can be stopped part-way through.
	ctx, cancel := cancellation()
	defer cancel()

	switch arguments.method {
		case "asset":
			err = GenerateAssetFile(&logger, arguments)

		case "manifest":
			err = GenerateManifestFile(&logger, arguments)

		case "check":
			err = Check(&logger, arguments)

		case "check-repo":
			err = CheckRepo(ctx, &logger, arguments)

		case "check-updates":
			err = CheckUpdates(ctx, &logger, arguments)

		case "install":
			err = Install(ctx, &logger, arguments)

		case "list":
			err = List(ctx, &logger, arguments)

		case "uninstall":
			err = Uninstall(ctx, &logger, arguments)

		case "update":
			err = Update(ctx, &logger, arguments)

		case "verify":
			err = Verify(ctx, &logger, arguments)
	}

	// A cancelled run is reported rather than failed: it stopped at a safe point and logs
	// how to resume. The exit code is still non-zero, so a caller - the bundled uninstaller,
	// which deletes the installation directory afterwards - can't mistake it for a successfully
	// completed run.
	if errors.Is(err, context.Canceled) {
		logger.log(interruption(arguments.method))
		os.Exit(1)
	}

	if err != nil {
		logger.fail(err)
		os.Exit(1)
	}
}
