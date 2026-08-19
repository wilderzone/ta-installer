package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

	// DXVK translates the game's Direct3D 9 calls to Vulkan, and the game is a 32-bit
	// program, so the driver that has to be there is the 32-bit one. The container the
	// runtime runs in supplies its own 32-bit system libraries, but not graphics drivers:
	// those have to match the running kernel, so they're taken from the host, and have to
	// be installed there for both architectures.
	drivers := surveyVulkanDrivers()
	switch {
		case !drivers.present:
			problems = append(problems, "no Vulkan driver was found")

		case !drivers.present32:
			problems = append(problems, "no 32-bit Vulkan driver was found, which is the one a 32-bit game needs")

		case !drivers.hardware32:
			problems = append(problems, "only software Vulkan rendering is available to 32-bit programs, which will be far too slow to play")
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

	if len(problems) == 0 {
		return true
	}

	logger.log(RUNTIME_CHECK_FAILED, strings.Join(problems, ", "))

	// The command is named, never run. Installing these means driving the system's package
	// manager as root, over a matrix of distributions and drivers where the wrong guess
	// damages the graphics stack rather than the game - and on an NVIDIA system the package
	// has to match the driver already installed. Steam, which has been solving this for
	// longer than anyone, also only tells the user what's missing.
	if !drivers.hardware32 {
		if command := driverInstructions(drivers, distribution()); command != "" {
			logger.log(RUNTIME_DRIVERS_HINT, command)
		}
	}

	return false
}

// VULKAN_ICD_DIRECTORIES are where a Vulkan driver registers itself. The ICD manifests are the
// only part of a Vulkan installation with a specified location, so they're the one thing that
// can be inspected without loading the loader.
var VULKAN_ICD_DIRECTORIES = []string{"/usr/share/vulkan/icd.d", "/usr/local/share/vulkan/icd.d", "/etc/vulkan/icd.d"}

// LIBRARY_DIRECTORIES_32 are the places the distributions disagree about putting 32-bit shared
// libraries: Debian's multiarch directories, Arch's lib32, and Fedora's plain /usr/lib. Nothing
// is trusted to the directory it was found in - every candidate is confirmed by its own ELF
// header - so listing one that holds 64-bit libraries elsewhere costs nothing.
var LIBRARY_DIRECTORIES_32 = []string{"/usr/lib/i386-linux-gnu", "/lib/i386-linux-gnu", "/usr/lib32", "/lib32", "/usr/lib"}

// vulkanDrivers is what the installed drivers say about this system's ability to run the game,
// which turns on two things a plain "is Vulkan installed" answer misses. The game is a 32-bit
// program, so the driver it needs is the 32-bit one; and a software rasteriser registers itself
// exactly like a real driver while being far too slow to play.
type vulkanDrivers struct {
	present    bool
	present32  bool
	hardware32 bool
	// Whether the drivers are NVIDIA's, whose 32-bit libraries are packaged separately from
	// Mesa's and have to match the driver already installed.
	nvidia bool
}

// vulkanManifest is the part of an ICD manifest that matters here: which library implements
// the driver. The loader resolves a bare name against the system library path, and anything
// with a separator in it against the manifest's own directory.
type vulkanManifest struct {
	ICD struct {
		LibraryPath string `json:"library_path"`
	} `json:"ICD"`
}

// surveyVulkanDrivers reports what the registered Vulkan drivers can do.
func surveyVulkanDrivers() vulkanDrivers {
	drivers := vulkanDrivers{}

	directories := append([]string{}, VULKAN_ICD_DIRECTORIES...)
	if data := dataDirectory(); data != "" {
		directories = append(directories, filepath.Join(data, "vulkan", "icd.d"))
	}

	for _, directory := range directories {
		entries, err := os.ReadDir(directory)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			drivers.inspect(filepath.Join(directory, entry.Name()))
		}
	}

	return drivers
}

// inspect folds one ICD manifest into what's known so far.
func (drivers *vulkanDrivers) inspect(manifest string) {
	contents, err := os.ReadFile(manifest)
	if err != nil {
		return
	}
	parsed := vulkanManifest{}
	if json.Unmarshal(contents, &parsed) != nil || parsed.ICD.LibraryPath == "" {
		return
	}
	drivers.present = true

	// lavapipe is Mesa's software rasteriser, and llvmpipe the OpenGL half of it. Both are
	// installed as a matter of course, so counting either as a driver that can run the game
	// is how a machine with no usable graphics at all passes this check.
	name := strings.ToLower(filepath.Base(manifest) + " " + parsed.ICD.LibraryPath)
	software := strings.Contains(name, "lvp") || strings.Contains(name, "llvmpipe")
	if strings.Contains(name, "nvidia") {
		drivers.nvidia = true
	}

	if is32BitLibrary(manifest, parsed.ICD.LibraryPath) {
		drivers.present32 = true
		if !software {
			drivers.hardware32 = true
		}
	}
}

// is32BitLibrary resolves an ICD manifest's library the way the Vulkan loader would, and
// reports whether what it found is 32-bit.
func is32BitLibrary(manifest string, library string) bool {
	if strings.Contains(library, string(os.PathSeparator)) {
		if !filepath.IsAbs(library) {
			library = filepath.Join(filepath.Dir(manifest), library)
		}
		return is32Bit(library)
	}

	for _, directory := range LIBRARY_DIRECTORIES_32 {
		if is32Bit(filepath.Join(directory, library)) {
			return true
		}
	}
	return false
}

// is32Bit reports whether a file is a 32-bit ELF binary. The fifth byte of an ELF header is
// its class, so nothing past the header has to be read - and reading it beats trusting the
// directory a library was found in, which every distribution names differently.
func is32Bit(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	const ELFCLASS32 = 1
	header := make([]byte, 5)
	if _, err := io.ReadFull(file, header); err != nil {
		return false
	}
	return string(header[0:4]) == "\x7fELF" && header[4] == ELFCLASS32
}

// distribution identifies the family this system belongs to, so that a remedy can be named in
// its own package manager's terms. Reading ID_LIKE as well as ID is what makes a derivative -
// Mint, Pop!_OS, Nobara, EndeavourOS - answer as the distribution it was built from.
func distribution() string {
	contents, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return ""
	}
	return distributionFrom(string(contents))
}

// distributionFrom picks the family out of an os-release file's contents.
func distributionFrom(contents string) string {
	identifiers := []string{}
	for _, line := range strings.Split(contents, "\n") {
		name, value, found := strings.Cut(strings.TrimSpace(line), "=")
		if !found || (name != "ID" && name != "ID_LIKE") {
			continue
		}
		identifiers = append(identifiers, strings.Fields(strings.Trim(value, `"`))...)
	}

	for _, identifier := range identifiers {
		switch identifier {
			case "debian", "ubuntu":
				return "debian"
			case "fedora", "rhel":
				return "fedora"
			case "arch":
				return "arch"
		}
	}
	return ""
}

// driverInstructions returns the command that installs 32-bit graphics drivers on the given
// distribution, or "" where there isn't one worth printing - an unrecognised distribution, or
// an NVIDIA system whose driver version can't be read, where the package name would be a guess
// that sends the user to something that doesn't exist.
//
// This is only ever printed. See preflight for why TAI doesn't run it.
func driverInstructions(drivers vulkanDrivers, family string) string {
	switch family {
		case "debian":
			packages := "libvulkan1:i386 mesa-vulkan-drivers:i386 libgl1-mesa-dri:i386"
			if drivers.nvidia {
				branch := nvidiaDriverBranch()
				if branch == "" {
					return ""
				}
				packages = "libvulkan1:i386 libnvidia-gl-" + branch + ":i386"
			}
			return "sudo dpkg --add-architecture i386 && sudo apt update && sudo apt install " + packages

		case "fedora":
			packages := "vulkan-loader.i686 mesa-vulkan-drivers.i686"
			if drivers.nvidia {
				packages = "vulkan-loader.i686 xorg-x11-drv-nvidia-libs.i686"
			}
			return "sudo dnf install " + packages

		case "arch":
			packages := "lib32-vulkan-icd-loader lib32-mesa lib32-vulkan-radeon lib32-vulkan-intel"
			if drivers.nvidia {
				packages = "lib32-vulkan-icd-loader lib32-nvidia-utils"
			}
			return "sudo pacman -S --needed " + packages
	}

	return ""
}

// nvidiaDriverBranch reads the major version of the loaded NVIDIA driver, which is what names
// its packages. The 32-bit libraries have to match the driver already installed, so this is
// read rather than assumed.
func nvidiaDriverBranch() string {
	version, err := os.ReadFile("/sys/module/nvidia/version")
	if err != nil {
		return ""
	}
	branch, _, _ := strings.Cut(strings.TrimSpace(string(version)), ".")
	return branch
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
