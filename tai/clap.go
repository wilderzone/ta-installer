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

// commandSpec describes which arguments a subcommand requires.
type commandSpec struct {
	needsPath  bool // Requires a valid absolute installation path.
	needsRepo  bool // Requires a valid repository URL (defaulted when empty).
	needsName  bool // Requires a valid asset name.
	pathOrRepo bool // Requires either a path or a repo (overrides the flags above).
}

var commands = map[string]commandSpec{
	"asset":      {needsPath: true, needsName: true},
	"manifest":   {needsPath: true},
	"list":       {pathOrRepo: true},
	"check":      {needsPath: true},
	"check-repo": {needsRepo: true},
	"install":    {needsPath: true, needsRepo: true},
	"verify":     {needsPath: true, needsRepo: true},
	"update":     {needsPath: true, needsRepo: true},
	"uninstall":  {needsPath: true},
}

// Guard against `methods` and `commands` drifting out of sync.
func init() {
	if len(methods) != len(commands) {
		panic("methods and commands are out of sync")
	}
	for _, method := range methods {
		if _, ok := commands[method]; !ok {
			panic("missing command spec for subcommand: " + method)
		}
	}
}

type arguments struct {
	method    string
	path      string
	repo      string
	name      string
	include   []string
	exclude   []string
	shortcuts bool
	machine   bool
	silent    bool
	yes       bool
}

// Write a usage summary to the given writer.
func usage(w io.Writer) {
	fmt.Fprint(w, "Usage: tai <subcommand> [flags]\n\n")
	fmt.Fprintf(w, "Subcommands:\n  %s\n\n", strings.Join(methods[:], ", "))
	fmt.Fprint(w, "Flags:\n")
	fmt.Fprint(w, "  --path=<path>            The installation directory.\n")
	fmt.Fprint(w, "  --repo=<url>             The repository URL to download from.\n")
	fmt.Fprint(w, "  --shortcuts=true|false   Create Desktop and Start Menu shortcuts for the game (when installing).\n")
	fmt.Fprint(w, "  -m                       Output machine-readable JSON.\n")
	fmt.Fprint(w, "  -s                       Suppress all output.\n")
	fmt.Fprint(w, "  -y                       Skip confirmation prompts.\n")
}

// parseList splits a comma-separated flag value into trimmed, non-empty terms.
// TrimSpace only trims the ends of each term, so an asset name with internal
// spaces (e.g. "Arx Novena") is preserved.
func parseList(value string) []string {
	terms := strings.Split(value, ",")
	for i := range terms {
		terms[i] = strings.TrimSpace(terms[i])
	}
	return slices.DeleteFunc(terms, func(term string) bool {
		return term == ""
	})
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
	fs.BoolVar(&args.shortcuts, "shortcuts", true, "")
	fs.BoolVar(&args.machine, "m", false, "")
	fs.BoolVar(&args.silent, "s", false, "")
	fs.BoolVar(&args.yes, "y", false, "")
	fs.Func("include", "", func(value string) error {
		args.include = parseList(value)
		return nil
	})
	fs.Func("exclude", "", func(value string) error {
		args.exclude = parseList(value)
		return nil
	})

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

// Validate a command that accepts either an installation path or a repo URL.
func validatePathOrRepo(arguments *arguments) error {
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
		return fmt.Errorf("The provided repo URL %q is not valid.", arguments.repo)
	} else {
		return errors.New("Provide an installation path or repository URL to list package contents.")
	}
}

func ValidateArguments(arguments *arguments) error {
	spec := commands[arguments.method]

	// Some commands accept either a path or a repo (e.g. `list`).
	if spec.pathOrRepo {
		return validatePathOrRepo(arguments)
	}

	// Path.
	if spec.needsPath {
		path, err := validatePath(arguments.path)
		if err != nil {
			return err
		}
		arguments.path = path
	}

	// Repo.
	if spec.needsRepo {
		if len(arguments.repo) == 0 {
			arguments.repo = DEFAULT_REPO
		}
		_, err := url.ParseRequestURI(arguments.repo)
		if err != nil {
			return errors.New("The provided repository URL is not valid.")
		}
	}

	// Name.
	if spec.needsName {
		match, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", arguments.name)
		if !match {
			return errors.New("The provided asset name is not valid. Please use only letters, numbers, and underscores.")
		}
	}

	return nil
}
