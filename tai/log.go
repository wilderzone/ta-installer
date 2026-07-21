package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
)

type Counter struct {
	logger *Logger
	mu     sync.Mutex
	index  int
	total  int
}

// next advances the counter and logs a progress line. It is safe to call from
// multiple goroutines; the lock also serialises the progress output.
func (c *Counter) next(message string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.index++
	progress := PROGRESS
	progress.Index = c.index
	progress.Total = c.total
	c.logger.log(progress, message)
}

type LogLine struct {
	Index   int    `json:"index,omitempty"`
	Total   int    `json:"total,omitempty"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

type Logger struct {
	machine bool
	silent  bool
}

var (
	ASSET_CREATED             = LogLine{Type: "ASSET_CREATED", Message: ""}
	ASSETS_ARE_DAMAGED        = LogLine{Type: "ASSETS_ARE_DAMAGED", Message: ""}
	CHECKING_ASSETS           = LogLine{Type: "CHECKING_ASSETS", Message: "Checking assets..."}
	CHECKING_FOR_UPDATES      = LogLine{Type: "CHECKING_FOR_UPDATES", Message: "Checking for updates..."}
	DOWNLOADING_MANIFEST      = LogLine{Type: "DOWNLOADING_MANIFEST", Message: "Downloading manifest from:"}
	INSTALL_FAILURE           = LogLine{Type: "INSTALL_FAILURE", Message: "Installation failed."}
	INSTALL_FOUND             = LogLine{Type: "INSTALL_FOUND", Message: "Installation found."}
	INSTALL_NOT_FOUND         = LogLine{Type: "INSTALL_NOT_FOUND", Message: "Installation not found."}
	INSTALL_PARTIAL           = LogLine{Type: "INSTALL_PARTIAL", Message: "Partially installed. You may need to run the \"verify\" command."}
	INSTALL_SUCCESS           = LogLine{Type: "INSTALL_SUCCESS", Message: "Installed."}
	INSTALL_UP_TO_DATE        = LogLine{Type: "INSTALL_UP_TO_DATE", Message: ""}
	INSTALL_VERIFIED          = LogLine{Type: "INSTALL_VERIFIED", Message: "Installation verified."}
	INSTALLING                = LogLine{Type: "INSTALLING", Message: "Installing..."}
	MANIFEST_CONTENTS         = LogLine{Type: "MANIFEST_CONTENTS", Message: ""}
	MANIFEST_CREATED          = LogLine{Type: "MANIFEST_CREATED", Message: ""}
	MANIFEST_UNREADABLE       = LogLine{Type: "MANIFEST_UNREADABLE", Message: ""}
	PATCHING_ASSETS           = LogLine{Type: "PATCHING_ASSETS", Message: "Patching assets..."}
	PREREQUISITES_DOWNLOADING = LogLine{Type: "PREREQUISITES_DOWNLOADING", Message: "Downloading DirectX 9..."}
	PREREQUISITES_FAILED      = LogLine{Type: "PREREQUISITES_FAILED", Message: "Failed to install DirectX 9:"}
	PREREQUISITES_INSTALLING  = LogLine{Type: "PREREQUISITES_INSTALLING", Message: "Installing DirectX 9..."}
	PREREQUISITES_REQUIRED    = LogLine{Type: "PREREQUISITES_REQUIRED", Message: "DirectX 9 needs to be installed before continuing."}
	PROMPT_PROCEED            = LogLine{Type: "PROMPT_PROCEED", Message: "Proceed? [y/N] "}
	PROMPT_UNINSTALL_ANYWAY   = LogLine{Type: "PROMPT_UNINSTALL_ANYWAY", Message: "Remove all files in this folder anyway? [y/N] "}
	READING_LOCAL_MANIFEST    = LogLine{Type: "READING_LOCAL_MANIFEST", Message: "Reading local manifest from:"}
	READING_MANIFEST_FAILED   = LogLine{Type: "READING_MANIFEST_FAILED", Message: "Failed to read manifest:"}
	REDOWNLOADING             = LogLine{Type: "REDOWNLOADING", Message: "Redownloading..."}
	REPO_FOUND                = LogLine{Type: "REPO_FOUND", Message: "Repository found."}
	REPO_NOT_FOUND            = LogLine{Type: "REPO_NOT_FOUND", Message: "Repository not found."}
	SHORTCUTS_CREATING        = LogLine{Type: "SHORTCUTS_CREATING", Message: "Creating shortcuts..."}
	SHORTCUTS_FAILURE         = LogLine{Type: "SHORTCUTS_FAILURE", Message: "Failed to create shortcuts. You may need to re-run the installer with admin permissions to create shortcuts."}
	SHORTCUTS_SUCCESS         = LogLine{Type: "SHORTCUTS_SUCCESS", Message: "Created shortcuts."}
	SHORTCUTS_REMOVE_FAILURE  = LogLine{Type: "SHORTCUTS_REMOVE_FAILURE", Message: "Failed to remove some shortcuts. You may need to delete them manually."}
	UNINSTALL_FAILURE         = LogLine{Type: "UNINSTALL_FAILURE", Message: "Failed to uninstall."}
	UNINSTALL_MATCHED_ALL     = LogLine{Type: "UNINSTALL_MATCHED_ALL", Message: "The provided filters match all assets. This will remove the entire installation."}
	UNINSTALL_PARTIAL         = LogLine{Type: "UNINSTALL_PARTIAL", Message: "Partially uninstalled. You may need to remove some files manually."}
	UNINSTALL_SUCCESS         = LogLine{Type: "UNINSTALL_SUCCESS", Message: "Uninstalled."}
	UNINSTALLING              = LogLine{Type: "UNINSTALLING", Message: "Uninstalling..."}
	UPDATE_FAILURE            = LogLine{Type: "UPDATE_FAILURE", Message: "Update failed."}
	UPDATE_FOUND              = LogLine{Type: "UPDATE_FOUND", Message: "Update found:"}
	UPDATE_PARTIAL            = LogLine{Type: "UPDATE_PARTIAL", Message: "Partially updated. You may need to run the \"verify\" command."}
	UPDATE_SUCCESS            = LogLine{Type: "UPDATE_SUCCESS", Message: "Updated."}
	PROGRESS                  = LogLine{Type: "PROGRESS", Index: 0, Total: 0, Message: ""}
	ERROR                     = LogLine{Type: "ERROR", Message: ""}
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
	return &Counter{logger: l, index: 0, total: total}
}

func (l *Logger) error(message ...any) {
	fmt.Fprintln(os.Stderr, message...)
}

// fail renders a top-level error for the user. It is the single place where
// errors are turned into user-facing output. Errors always go to stderr, keeping
// stdout clean for data and progress logs.
func (l *Logger) fail(err error) {
	if l.machine {
		// Structured error line on stderr, so it never corrupts stdout data.
		line := ERROR
		line.Message = err.Error()
		json.NewEncoder(os.Stderr).Encode(line)
		return
	}

	var userError *UserError
	if errors.As(err, &userError) {
		l.error(l.colour(userError.Message, RED))
	} else {
		l.error(l.colour("Something went wrong: ", RED) + err.Error())
	}
}

func (l *Logger) log(message LogLine, additional ...any) {
	if l.silent {
		return
	}

	message.Message = strings.TrimSpace(message.Message + " " + fmt.Sprint(additional...))

	if l.machine {
		encoder := json.NewEncoder(os.Stdout)
		err := encoder.Encode(message)
		if err != nil {
			l.error(err.Error())
		}
	} else {
		if message.Type == "PROGRESS" {
			counter := l.colour(fmt.Sprintf("[%d/%d] ", message.Index, message.Total), GREY)
			message.Message = counter + message.Message
		}
		fmt.Println(message.Message)
	}
}

func (l *Logger) plural(count int, singular string, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func (l *Logger) prompt(message LogLine) bool {
	line := ""
	if l.machine {
		line = message.Type
	} else {
		line = message.Message
	}
	fmt.Print(line)
	action := "n"
	_, _ = fmt.Scanln(&action)
	match, _ := regexp.MatchString("^[yY]", action)
	return match
}
