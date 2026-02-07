package main

import (
	"errors"
	"io/fs"
	"net/url"
	"os"
	"regexp"
	"slices"
)

var methods = [...]string{
	"asset",
	"manifest",
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
}

func GetArguments() arguments {
	method := ""
	path := ""
	repo := ""

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
	}

	return arguments{
		method: method,
		path:   path,
		repo:   repo,
	}
}

func ValidateArguments(arguments arguments) error {
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
		_, err := url.ParseRequestURI(arguments.repo)
		if err != nil {
			return errors.New("The provided repository URL is not valid.")
		}
	}

	return nil
}
