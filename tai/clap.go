package main

import (
	"errors"
	"io/fs"
	"net/url"
	"os"
	"regexp"
	"slices"
)

const DEFAULT_REPO = "https://cdn.wilderzone.live/ta"

var methods = [...]string{
	"asset",
	"manifest",
	"check",
	"install",
	"verify",
	"patch",
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

func GetArguments() arguments {
	method := ""
	path := ""
	repo := ""
	name := ""
	machine := false
	silent := false
	yes := false

	for _, argument := range os.Args[1:] {
		// Method.
		if len(method) == 0 && slices.Contains(methods[:], argument) {
			method = argument
			continue
		}
		// Path.
		match, _ := regexp.MatchString("^--path=", argument)
		if len(path) == 0 && match {
			path = argument[7:]
			continue
		}
		// Repo.
		match, _ = regexp.MatchString("^--repo=", argument)
		if len(repo) == 0 && match {
			repo = argument[7:]
			continue
		}
		// Name.
		match, _ = regexp.MatchString("^--name=", argument)
		if len(name) == 0 && match {
			name = argument[7:]
			continue
		}
		// Yes shortcut.
		match, _ = regexp.MatchString("^-y$", argument)
		if match {
			yes = true
			continue
		}
		// Machine-readable flag.
		match, _ = regexp.MatchString("^-m$", argument)
		if match {
			machine = true
			continue
		}
		// Silent flag.
		match, _ = regexp.MatchString("^-s$", argument)
		if match {
			silent = true
			yes = true
			continue
		}
	}

	return arguments{
		method:  method,
		path:    path,
		repo:    repo,
		name:    name,
		machine: machine,
		silent:  silent,
		yes:     yes,
	}
}

func ValidateArguments(arguments *arguments) error {
	// Method.
	if !slices.Contains(methods[:], arguments.method) {
		return errors.New("No subcommand provided.")
	}

	// Path.
	if !fs.ValidPath(arguments.path) {
		return errors.New("The provided path is not valid.")
	}

	// Repo.
	if arguments.method != "asset" && arguments.method != "manifest" {
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
		match, _ := regexp.MatchString("[a-zA-Z0-9_]+", arguments.name)
		if !match {
			return errors.New("The provided asset name is not valid. Please use only letters, numbers, and underscores.")
		}
	}

	return nil
}
