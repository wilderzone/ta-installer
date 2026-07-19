package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// The exit code Windows reports when the user dismisses the UAC prompt.
const ERROR_CANCELLED = 1223

// ELEVATE_SCRIPT launches an elevated process and reports its exit code.
//
// This uses [Diagnostics.Process]::Start rather than the more obvious
// Start-Process cmdlet, because Start-Process rewrites the underlying
// Win32Exception into an InvalidOperationException with no inner exception,
// discarding the numeric error code and leaving it only in localised message
// text. Starting the process directly preserves NativeErrorCode, which is what
// lets a dismissed UAC prompt be told apart from a genuine failure.
//
// The null check matters too: on failure $process is left null, and a bare
// "exit $process.ExitCode" exits 0, reporting success for a launch that never
// happened.
const ELEVATE_SCRIPT = `
$ErrorActionPreference = 'Stop'
$info = New-Object System.Diagnostics.ProcessStartInfo
$info.FileName = %s
$info.Arguments = %s
$info.Verb = 'runas'
$info.UseShellExecute = $true
$info.WindowStyle = 'Hidden'
try {
	$process = [System.Diagnostics.Process]::Start($info)
} catch [System.ComponentModel.Win32Exception] {
	exit $_.Exception.NativeErrorCode
} catch {
	exit 1
}
if ($null -eq $process) { exit 1 }
$process.WaitForExit()
exit $process.ExitCode
`

// quoteForPowerShell wraps a string as a PowerShell single-quoted literal, in
// which the only escape is a doubled quote. Single quotes prevent PowerShell
// from expanding "$" and other metacharacters in the value.
func quoteForPowerShell(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

// quoteForCommandLine quotes a single argument for a Windows command line, which
// the elevated process parses itself. This is a separate layer from the
// PowerShell quoting applied afterwards.
func quoteForCommandLine(value string) string {
	if value != "" && !strings.ContainsAny(value, " \t\"") {
		return value
	}
	return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
}

// runElevated runs the given executable with administrator privileges, prompting
// the user for consent via UAC, and blocks until it exits.
//
// Elevation is granted at process-creation time by the "runas" verb, which
// os/exec cannot request, so this delegates to PowerShell. The elevated process
// runs on a separate security desktop and its output cannot be captured - only
// its exit code is available.
func runElevated(path string, arguments ...string) error {
	quoted := make([]string, len(arguments))
	for index, argument := range arguments {
		quoted[index] = quoteForCommandLine(argument)
	}
	script := fmt.Sprintf(ELEVATE_SCRIPT,
		quoteForPowerShell(path),
		quoteForPowerShell(strings.Join(quoted, " ")),
	)

	command := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	err := command.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == ERROR_CANCELLED {
			return userErrorf(err, "DirectX 9 wasn't installed because administrator permission was declined. Run the installer again and accept the prompt to finish setting up.")
		}
		return fmt.Errorf("run %s as administrator: %w", path, err)
	}

	return nil
}
