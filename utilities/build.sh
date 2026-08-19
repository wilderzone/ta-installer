#! /bin/bash

SCRIPT_CWD="$(dirname "$(readlink -f "$0")")"
SCRIPT_LINUX_SUFFIX="-linux-amd64"
SCRIPT_WINDOWS_SUFFIX="-windows-amd64"

build() {
	local PLATFORM=$1
	local EXTENSION=$2
	local OUT="$SCRIPT_CWD/dist/$PLATFORM"

	echo $PLATFORM

	rm -rf $OUT

	# Build shared utilities.
	echo "- CLI"
	cd "$SCRIPT_CWD/../tai" && go build -o "$OUT/tai-cli$EXTENSION" .
	chmod +x "$OUT/tai-cli$EXTENSION"

	echo "- play"
	if [[ $PLATFORM == "Linux" ]]; then
		cd "$SCRIPT_CWD/play"   && go build -o "$OUT/play$EXTENSION" .
	fi
	if [[ $PLATFORM == "Windows" ]]; then
		cd "$SCRIPT_CWD/play"   && go build -ldflags="-H=windowsgui" -o "$OUT/play$EXTENSION" .
	fi
	chmod +x "$OUT/play$EXTENSION"

	## Build Windows-specific utilities.
	if [[ $PLATFORM == "Windows" ]]; then
		echo "- uninstall"
		cd "$SCRIPT_CWD/uninstall" && go build -o "$OUT/uninstall$EXTENSION" .
		chmod +x "$OUT/uninstall$EXTENSION"
		echo "- update"
		cd "$SCRIPT_CWD/update"    && go build -o "$OUT/update$EXTENSION" .
		chmod +x "$OUT/update$EXTENSION"
		echo "- verify"
		cd "$SCRIPT_CWD/verify"    && go build -o "$OUT/verify$EXTENSION" .
		chmod +x "$OUT/verify$EXTENSION"
	fi

	# Copy the icon files.
	cp "$SCRIPT_CWD/icon.png" "$OUT/icon.png"
	cp "$SCRIPT_CWD/uninstall.png" "$OUT/uninstall.png"

	# Generate the asset file and zip the package.
	$SCRIPT_CWD/dist/Linux/tai-cli asset --path="$OUT" --name=Installer
	if [[ $PLATFORM == "Linux" ]]; then
		mv "$OUT/Installer.json" "$OUT/Installer$SCRIPT_LINUX_SUFFIX.json"
		cd "$OUT" && zip -9 "Installer$SCRIPT_LINUX_SUFFIX.zip" "icon.png" "uninstall.png" "tai-cli$EXTENSION" "play$EXTENSION"
	fi
	if [[ $PLATFORM == "Windows" ]]; then
		mv "$OUT/Installer.json" "$OUT/Installer$SCRIPT_WINDOWS_SUFFIX.json"
		cd "$OUT" && zip -9 "Installer$SCRIPT_WINDOWS_SUFFIX.zip" "tai-cli$EXTENSION" "play$EXTENSION" "uninstall$EXTENSION" "update$EXTENSION" "verify$EXTENSION"
	fi
}

echo ""

export GOOS=linux
build "Linux" ""
echo ""

export GOOS=windows
build "Windows" ".exe"
echo ""
