package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"wilderzone.org/tai/internal/assets"
)

// The compatibility runtime lives inside the installation, so that uninstalling the game
// is still "remove one directory" and so that each installation pins the runtime it was
// tested against.
const RUNTIME_DIRECTORY = "Runtime"
const PREFIX_DIRECTORY = "Prefix"
const UMU_DIRECTORY = "umu"
const UMU_EXECUTABLE = "umu-run"

// The identifier umu uses to look up per-game fixes. Tribes: Ascend's Steam application ID
// opts the standalone build into the same Proton fixes the Steam release gets. If those fixes
// ever do something the standalone build dislikes, "umu-default" turns them off.
const UMU_GAME_ID = "umu-17080"

// runtimeDirectory returns the directory the compatibility runtime is unpacked into.
func runtimeDirectory(root string) string {
	return filepath.Join(root, RUNTIME_DIRECTORY)
}

// prefixDirectory returns the WINEPREFIX for this installation.
func prefixDirectory(root string) string {
	return filepath.Join(root, PREFIX_DIRECTORY)
}

// umuExecutable returns the path to the launcher shim.
func umuExecutable(root string) string {
	return filepath.Join(runtimeDirectory(root), UMU_DIRECTORY, UMU_EXECUTABLE)
}

// protonDirectory finds the unpacked Proton build inside the runtime directory. The build is
// located by looking for the "proton" entry point rather than by name, so that publishing a
// newer GE-Proton doesn't need a matching change here.
func protonDirectory(root string) (string, error) {
	entries, err := os.ReadDir(runtimeDirectory(root))
	if err != nil {
		return "", fmt.Errorf("read the runtime directory: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidate := filepath.Join(runtimeDirectory(root), entry.Name())
		if fileExists(filepath.Join(candidate, "proton")) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("no Proton build found in %s", runtimeDirectory(root))
}

// runtimeIsMissing reports whether the pieces the game needs to run are absent. It mirrors
// installationUtilitiesAreMissing: cheap, and only ever used to decide whether to set things
// up, never to declare an installation broken.
func runtimeIsMissing(root string) bool {
	if !fileExists(umuExecutable(root)) {
		return true
	}
	_, err := protonDirectory(root)
	return err != nil
}

// windowsPath converts an absolute Unix path to the form the game sees from inside the
// prefix. Wine maps the whole filesystem onto the Z: drive, so no prefix surgery is needed
// to make an installation outside the prefix reachable.
func windowsPath(path string) string {
	return "Z:" + strings.ReplaceAll(path, "/", `\`)
}

// runtimeCommand builds a command that runs a Windows executable inside this installation's
// prefix. Everything umu needs is passed through the environment; nothing is written to the
// user's shell configuration or to a wrapper script.
func runtimeCommand(ctx context.Context, root string, executable string, args ...string) (*exec.Cmd, error) {
	proton, err := protonDirectory(root)
	if err != nil {
		return nil, err
	}

	command := exec.CommandContext(ctx, umuExecutable(root), append([]string{executable}, args...)...)
	command.Env = append(os.Environ(),
		"GAMEID=" + UMU_GAME_ID,
		"STORE=none",
		"PROTONPATH=" + proton,
		"WINEPREFIX=" + prefixDirectory(root),
		"PROTON_VERB=waitforexitandrun",
		// The runtime is pinned by the manifest and verified by checksum, so leave it
		// alone rather than letting umu check for a newer one on every launch.
		"UMU_RUNTIME_UPDATE=0",
	)
	return command, nil
}

// downloadRuntime fetches the compatibility runtime, unless it's already installed. It
// mirrors downloadInstallationUtilities: always from DEFAULT_REPO, never from --repo, because
// the runtime is TAI's own infrastructure rather than something a game repository provides.
//
// Unlike the utilities it's measured, because half a gigabyte behind a bare spinner reads as
// a hang.
func downloadRuntime(ctx context.Context, logger *Logger, root string) error {
	if !runtimeIsMissing(root) {
		return nil
	}

	logger.log(RUNTIME_DOWNLOADING)
	counter := logger.count(1, "Downloading")
	defer counter.done()
	progress := counter.measure(COMPATIBILITY_RUNTIME.PackedSize)

	err := download(ctx, DEFAULT_REPO, root, COMPATIBILITY_RUNTIME, progress)
	if err != nil {
		if cancelled := ctx.Err(); cancelled != nil {
			return cancelled
		}
		counter.next(COMPATIBILITY_RUNTIME.Name + ": " + logger.colour("Failed", RED))
		return userErrorf(err, "Couldn't download the compatibility layer, so the game won't be able to run.")
	}

	counter.next(COMPATIBILITY_RUNTIME.Name + ": " + logger.colour("Downloaded", GREEN))
	return nil
}

// verifyRuntime checks the installed runtime against the checksums published on the official
// CDN, and replaces it if the user agrees. Like verifyUtilities it never returns an error:
// the game files have already been verified by the time it runs, so a missing network is
// reported and swallowed rather than failing the whole command.
func verifyRuntime(ctx context.Context, logger *Logger, arguments arguments) {
	if runtimeIsMissing(arguments.path) {
		logger.log(RUNTIME_MISSING)
		return
	}

	asset, err := getPublishedAsset(ctx, DEFAULT_REPO, RUNTIME_CHECKSUMS_SOURCE, COMPATIBILITY_RUNTIME)
	if err != nil {
		// A cancelled run reports the interruption itself; this wasn't skipped.
		if ctx.Err() == nil {
			logger.log(RUNTIME_CHECK_SKIPPED, err.Error())
		}
		return
	}

	damaged := assets.VerifyAsset(ctx, arguments.path, asset)
	if ctx.Err() != nil {
		return
	}
	if len(damaged) == 0 {
		logger.log(RUNTIME_VERIFIED)
		return
	}
	logger.log(RUNTIME_DAMAGED, assets.Describe(damaged))

	proceed := arguments.yes || logger.prompt(PROMPT_PROCEED)
	if !proceed {
		return
	}

	// The runtime is replaced wholesale rather than file by file. It's a tree of symlinks
	// and hardlinks that only makes sense as a unit, and removing it first is also what
	// keeps a newly pinned build from being unpacked alongside the one it replaces.
	err = os.RemoveAll(runtimeDirectory(arguments.path))
	if err != nil {
		logger.detail("remove the old compatibility runtime", err)
		return
	}
	err = downloadRuntime(ctx, logger, arguments.path)
	if err != nil {
		logger.detail("replace the compatibility runtime", err)
		return
	}
	logger.log(RUNTIME_REPAIRED)
}

// ensureRuntime prepares the compatibility layer: it confirms the runtime is present, then
// creates the prefix if it isn't there yet.
//
// Creating the prefix is also what makes umu fetch its container image, which it keeps in
// the user's data directory rather than in the installation. That download is outside TAI's
// own progress reporting, so it runs behind a spinner with umu's own output going to the log.
func ensureRuntime(ctx context.Context, logger *Logger, root string) error {
	if runtimeIsMissing(root) {
		return userErrorf(nil, "The compatibility layer wasn't installed, so the game can't run. Run \"verify\" to restore it.")
	}

	// A prefix that already has a system drive has been through this once already.
	if fileExists(filepath.Join(prefixDirectory(root), "pfx", "drive_c", "windows")) {
		return nil
	}

	logger.log(RUNTIME_PREPARING)
	err := os.MkdirAll(prefixDirectory(root), os.FileMode(0755))
	if err != nil {
		return fmt.Errorf("create the prefix directory: %w", err)
	}

	// wineboot only initialises the prefix and exits, which is all this needs; the point
	// is the setup umu does around it.
	command, err := runtimeCommand(ctx, root, "wineboot", "-u")
	if err != nil {
		return err
	}
	spinner := logger.spin("")
	output, err := command.CombinedOutput()
	spinner.stop()
	if cancelled := ctx.Err(); cancelled != nil {
		return cancelled
	}
	if err != nil {
		logger.detail("prepare the compatibility layer", fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output))))
		return userErrorf(err, "Couldn't prepare the compatibility layer. Check that this system can run containers, then try again.")
	}

	logger.log(RUNTIME_READY)
	return nil
}

// launchGame starts the game inside the compatibility layer.
func launchGame(ctx context.Context, logger *Logger, root string) error {
	if runtimeIsMissing(root) {
		logger.log(RUNTIME_MISSING)
		return userErrorf(nil, "The compatibility runtime is missing, so the game can't start.")
	}

	working := filepath.Join(root, "Binaries", "Win32")
	logs := filepath.Join(root, LOGS_DIRECTORY, LOG_NAME)
	err := os.MkdirAll(filepath.Dir(logs), os.FileMode(0755))
	if err != nil {
		return fmt.Errorf("create the logs directory: %w", err)
	}

	// The game reads -abslog as a Windows path, so it has to be given one even though
	// everything around it here is a Unix path.
	arguments := strings.Fields(fmt.Sprintf(SHORTCUT_ARGUMENTS, windowsPath(logs)))
	command, err := runtimeCommand(ctx, root, filepath.Join(working, GAME_BINARY_NAME), arguments...)
	if err != nil {
		return err
	}
	command.Dir = working
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	logger.log(LAUNCHING)
	return command.Run()
}

// preflight reports anything about this machine that would stop the compatibility layer
// working. These are findings rather than errors: they're worth telling the user about
// before a multi-gigabyte download, but none of them is worth refusing to install over.
func preflight(logger *Logger) {
	problems := []string{}

	// DXVK translates the game's Direct3D 9 calls to Vulkan. Without a Vulkan driver the
	// runtime falls back to OpenGL, which is much slower.
	if !hasVulkanDriver() {
		problems = append(problems, "no Vulkan driver was found")
	}

	// umu is distributed as a Python zipapp.
	if _, err := exec.LookPath("python3"); err != nil {
		problems = append(problems, "python3 isn't installed")
	}

	// The container the runtime runs in needs unprivileged user namespaces, which some
	// hardened kernels switch off.
	if !hasUserNamespaces() {
		problems = append(problems, "unprivileged user namespaces are disabled")
	}

	if len(problems) != 0 {
		logger.log(RUNTIME_CHECK_FAILED, strings.Join(problems, ", "))
	}
}

// hasVulkanDriver reports whether any Vulkan driver has registered itself. The ICD manifests
// are the only part of a Vulkan installation with a specified location, so they're the one
// thing that can be checked without loading the loader.
func hasVulkanDriver() bool {
	directories := []string{"/usr/share/vulkan/icd.d", "/usr/local/share/vulkan/icd.d", "/etc/vulkan/icd.d"}
	if data := dataDirectory(); data != "" {
		directories = append(directories, filepath.Join(data, "vulkan", "icd.d"))
	}
	for _, directory := range directories {
		entries, err := os.ReadDir(directory)
		if err == nil && len(entries) != 0 {
			return true
		}
	}
	return false
}

// hasUserNamespaces reports whether this kernel lets an unprivileged process create a user
// namespace. Debian and Ubuntu expose a switch for it; everywhere else the absence of the
// switch means the feature is simply on.
func hasUserNamespaces() bool {
	value, err := os.ReadFile("/proc/sys/kernel/unprivileged_userns_clone")
	if err != nil {
		return true
	}
	return strings.TrimSpace(string(value)) != "0"
}
