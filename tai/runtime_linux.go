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
// tested against. What it takes to *run* the game through it belongs to the play utility;
// the installer only puts it there and keeps it healthy.
const RUNTIME_DIRECTORY = "Runtime"

// runtimeDirectory returns the directory the compatibility runtime is unpacked into.
func runtimeDirectory(root string) string {
	return filepath.Join(root, RUNTIME_DIRECTORY)
}

// runtimeIsMissing reports whether the compatibility runtime is absent. It mirrors
// installationUtilitiesAreMissing: cheap, and only ever used to decide whether to fetch it,
// never to declare an installation broken.
func runtimeIsMissing(root string) bool {
	for _, file := range COMPATIBILITY_RUNTIME.Files {
		if !fileExists(filepath.Join(root, COMPATIBILITY_RUNTIME.Destination, file.File)) {
			return true
		}
	}
	return false
}

// downloadRuntime fetches the compatibility runtime, unless it's already installed. It
// mirrors downloadInstallationUtilities: always from DEFAULT_REPO, never from --repo, because
// the runtime is TAI's own infrastructure rather than something a game repository provides.
//
// Unlike the utilities, it's measured, because half a gigabyte behind a bare spinner looks like
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
		return userErrorf(err, "Couldn't download the compatibility layer, so the game won't be able to start.")
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
		logger.detail("remove the old compatibility layer", err)
		return
	}
	err = downloadRuntime(ctx, logger, arguments.path)
	if err != nil {
		logger.detail("replace the compatibility layer", err)
		return
	}
	logger.log(RUNTIME_REPAIRED)
}

// ensureRuntime has the compatibility layer prepared, by asking the play utility to do it.
//
// The launch contract - which Proton build, which environment, which prefix - belongs to the
// launcher, so the installer delegates rather than keeping a second copy of it that could
// drift. Preparing it now rather than on first launch matters because it is also when umu
// fetches its container image: several hundred megabytes that would otherwise be paid for
// when the user expects the game to start, with nothing on screen to explain the wait.
//
// A failure here is reported but not fatal. The runtime is installed either way, and the
// launcher prepares the prefix itself on first launch if this didn't get to it.
func ensureRuntime(ctx context.Context, logger *Logger, root string) error {
	launcher := filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY, PLAY_NAME)
	if !fileExists(launcher) {
		logger.detail("prepare the compatibility layer", fmt.Errorf("%s isn't installed", PLAY_NAME))
		return nil
	}

	logger.log(RUNTIME_PREPARING)
	command := exec.CommandContext(ctx, launcher, "-prepare")
	spinner := logger.spin("")
	output, err := command.CombinedOutput()
	spinner.stop()
	if cancelled := ctx.Err(); cancelled != nil {
		return cancelled
	}
	if err != nil {
		logger.detail("prepare the compatibility layer", fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output))))
		logger.log(RUNTIME_PREPARE_FAILED)
		return nil
	}

	logger.log(RUNTIME_READY)
	return nil
}

// preflight reports anything about this machine that would stop the compatibility
// layer working, before TAI commits to a multi-gigabyte download.
//
// Returns true if the preflight check was successful.
func preflight(logger *Logger) bool {
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
		return false
	}

	return true
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
