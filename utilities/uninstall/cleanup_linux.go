package main

import "os"

// scheduleCleanup removes what the CLI had to leave behind. On Windows the final step has to
// be deferred to a detached script, because the utilities directory still holds this running
// executable. Unlinking a running executable is fine here, so the removal happens directly.
func scheduleCleanup(game string) {
	_ = os.RemoveAll(game)
}
