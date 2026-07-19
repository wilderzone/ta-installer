package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

const DEFAULT_REPO = "https://cdn.wilderzone.live/ta"

var methods = [...]string{
	"asset",
	"manifest",
	"list",
	"check",
	"check-repo",
	"install",
	"verify",
	"update",
	"uninstall",
}

type arguments struct {
	method  string
	path    string
	repo    string
	name    string
	machine bool
	silent  bool
	yes     bool
}

// Write a usage summary to the given writer.
func usage(w io.Writer) {
	fmt.Fprint(w, "Usage: tai <subcommand> [flags]\n\n")
	fmt.Fprintf(w, "Subcommands:\n  %s\n\n", strings.Join(methods[:], ", "))
	fmt.Fprint(w, "Flags:\n")
	fmt.Fprint(w, "  --path=<path>  The installation directory.\n")
	fmt.Fprint(w, "  --repo=<url>   The repository URL to download from.\n")
	fmt.Fprint(w, "  -m             Output machine-readable JSON.\n")
	fmt.Fprint(w, "  -s             Suppress all output.\n")
	fmt.Fprint(w, "  -y             Skip confirmation prompts.\n")
}

func GetArguments() (arguments, error) {
	args := arguments{}

	// The subcommand must be the first argument.
	if len(os.Args) < 2 {
		return args, errors.New("No subcommand provided.")
	}
	args.method = os.Args[1]
	if !slices.Contains(methods[:], args.method) {
		return args, fmt.Errorf("Unknown subcommand %q.", args.method)
	}

	// Parse the remaining flags. Repeated flags take their last occurrence.
	fs := flag.NewFlagSet(args.method, flag.ContinueOnError)
	fs.SetOutput(io.Discard) // Errors are surfaced to the caller instead.
	fs.StringVar(&args.path, "path", "", "")
	fs.StringVar(&args.repo, "repo", "", "")
	fs.StringVar(&args.name, "name", "", "")
	fs.BoolVar(&args.machine, "m", false, "")
	fs.BoolVar(&args.silent, "s", false, "")
	fs.BoolVar(&args.yes, "y", false, "")

	err := fs.Parse(os.Args[2:])
	if err != nil {
		// `-h`/`--help` is not a failure: print usage and exit cleanly.
		if errors.Is(err, flag.ErrHelp) {
			usage(os.Stdout)
			os.Exit(0)
		}
		return args, err
	}

	// Reject stray positional arguments (e.g. `tai install extrajunk`).
	if fs.NArg() > 0 {
		return args, fmt.Errorf("Unexpected argument %q.", fs.Arg(0))
	}

	// Silent mode can't answer prompts, so it implies -y.
	if args.silent {
		args.yes = true
	}

	return args, nil
}

// Validate that `path` is an absolute path for the current OS, and normalize it.
func validatePath(path string) (string, error) {
	if !filepath.IsAbs(path) {
			return "", fmt.Errorf("The provided path %q must be absolute.", path)
	}
	return filepath.Clean(path), nil
}

func ValidateArguments(arguments *arguments) error {
	// List subcommand - validate path and repo together.
	if arguments.method == "list" {
		path, pathErr := validatePath(arguments.path)
		if pathErr == nil {
			arguments.path = path
			return nil
		}
		_, repoErr := url.ParseRequestURI(arguments.repo)
		if repoErr == nil {
			return nil
		}
		if len(arguments.path) != 0 {
			return pathErr
		} else if len(arguments.repo) != 0 {
			return repoErr
		} else {
			return errors.New("Provide an installation path or repository URL to list package contents.")
		}
	}

	// Path.
	if arguments.method != "check-repo" {
		path, err := validatePath(arguments.path)
		if err != nil {
			return err
		}
		arguments.path = path
	}

	// Repo.
	if arguments.method != "asset" && arguments.method != "manifest" && arguments.method != "check" && arguments.method != "uninstall" && arguments.method != "verify" {
		if len(arguments.repo) == 0 {
			arguments.repo = DEFAULT_REPO
		}
		_, err := url.ParseRequestURI(arguments.repo)
		if err != nil {
			return errors.New("The provided repository URL is not valid.")
		}
	}

	// Name.
	if arguments.method == "asset" {
		match, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", arguments.name)
		if !match {
			return errors.New("The provided asset name is not valid. Please use only letters, numbers, and underscores.")
		}
	}

	return nil
}
