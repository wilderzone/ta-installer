package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Counter struct {
	logger  *Logger
	index   int
	total   int
	machine bool
}

func (c *Counter) next(message string) {
	c.index++
	m := ""
	if c.machine {
		m += fmt.Sprintf("[%d/%d] ", c.index, c.total)
	} else {
		m += c.logger.colour(fmt.Sprintf("[%d/%d] ", c.index, c.total), GREY)
	}
	m += message
	c.logger.log(PROGRESS, m)
}

type LogLine struct {
	human   string
	machine string
}

type Logger struct {
	machine bool
	silent  bool
}

var (
	ASSET_CREATED            = LogLine{machine: "ASSET_CREATED:", human: ""}
	ASSETS_ARE_DAMAGED       = LogLine{machine: "ASSETS_ARE_DAMAGED:", human: ""}
	CHECKING_ASSETS          = LogLine{machine: "CHECKING_ASSETS", human: "Checking assets..."}
	CHECKING_FOR_UPDATES     = LogLine{machine: "CHECKING_FOR_UPDATES", human: "Checking for updates..."}
	DOWNLOADING_MANIFEST     = LogLine{machine: "DOWNLOADING_MANIFEST:", human: "Downloading manifest from:"}
	INSTALL_FAILURE          = LogLine{machine: "INSTALL_FAILURE", human: "Installation failed."}
	INSTALL_PARTIAL          = LogLine{machine: "INSTALL_PARTIAL", human: "Partially installed. You may need to run the \"verify\" command."}
	INSTALL_SUCCESS          = LogLine{machine: "INSTALL_SUCCESS", human: "Installed."}
	INSTALL_UP_TO_DATE       = LogLine{machine: "INSTALL_UP_TO_DATE", human: ""}
	INSTALL_VERIFIED         = LogLine{machine: "INSTALL_VERIFIED", human: "Installation verified."}
	INSTALLING_PREREQUISITES = LogLine{machine: "INSTALLING_PREREQUISITES", human: "Installing prerequisites..."}
	MANIFEST_CREATED         = LogLine{machine: "MANIFEST_CREATED:", human: ""}
	PATCHING_ASSETS          = LogLine{machine: "PATCHING_ASSETS", human: "Patching assets..."}
	PROGRESS                 = LogLine{machine: "PROGRESS:", human: ""}
	PROMPT_PROCEED           = LogLine{machine: "PROMPT_PROCEED", human: "Proceed? [y/N] "}
	READING_LOCAL_MANIFEST   = LogLine{machine: "READING_LOCAL_MANIFEST:", human: "Reading local manifest from:"}
	REDOWNLOADING            = LogLine{machine: "REDOWNLOADING", human: "Redownloading..."}
	RETRIEVING_MANIFEST      = LogLine{machine: "RETRIEVING_MANIFEST:", human: "Retrieving manifest from:"}
	UPDATE_FAILURE           = LogLine{machine: "UPDATE_FAILURE", human: "Update failed."}
	UPDATE_FOUND             = LogLine{machine: "UPDATE_FOUND:", human: "Update found:"}
	UPDATE_PARTIAL           = LogLine{machine: "UPDATE_PARTIAL", human: "Partially updated. You may need to run the \"verify\" command."}
	UPDATE_SUCCESS           = LogLine{machine: "UPDATE_SUCCESS", human: "Updated."}
)

const (
	RESET = "\x1b[0m"
	GREY  = "\x1b[2m"
	GREEN = "\x1b[32m"
	RED   = "\x1b[31m"
)

func (l *Logger) colour(message string, colour string) string {
	if l.machine {
		return message
	}
	return colour + message + RESET
}

func (l *Logger) count(total int) *Counter {
	return &Counter{logger: l, index: 0, total: total, machine: l.machine}
}

func (l *Logger) error(message ...any) {
	fmt.Fprintln(os.Stderr, message...)
}

func (l *Logger) log(message LogLine, additional ...any) {
	if l.silent {
		return
	}

	line := ""
	if l.machine {
		line = message.machine
	} else {
		line = message.human
	}
	line = strings.TrimSpace(line + " " + fmt.Sprint(additional...))
	fmt.Println(line)
}

func (l *Logger) plural(count int, singular string, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func (l *Logger) print(message ...any) {
	if l.silent {
		return
	}

	fmt.Println(message...)
}

func (l *Logger) prompt(message LogLine) bool {
	line := ""
	if l.machine {
		line = message.machine
	} else {
		line = message.human
	}
	fmt.Print(line)
	action := "n"
	_, _ = fmt.Scanln(&action)
	match, _ := regexp.MatchString("^[yY]", action)
	return match
}
