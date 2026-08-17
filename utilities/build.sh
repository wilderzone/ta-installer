#! /bin/bash

SCRIPT_CWD="$(dirname "$(readlink -f "$0")")"
SCRIPT_LINUX_SUFFIX="-linux-amd64"

build() {
	local PLATFORM=$1
	local EXTENSION=$2

	echo ""
	echo $PLATFORM

	rm -r "$SCRIPT_CWD/dist/$PLATFORM"

	# Build shared utilities.
	echo "- CLI"
	cd "$SCRIPT_CWD/../tai" && go build -o "$SCRIPT_CWD/dist/$PLATFORM/tai-cli$EXTENSION" .
	chmod +x "$SCRIPT_CWD/dist/$PLATFORM/tai-cli$EXTENSION"

	if [[ $PLATFORM == "Linux" ]]; then
		echo "- play"
		cd "$SCRIPT_CWD/play"   && go build -o "$SCRIPT_CWD/dist/$PLATFORM/play$EXTENSION" .
		chmod +x "$SCRIPT_CWD/dist/$PLATFORM/play$EXTENSION"
	fi
	if [[ $PLATFORM == "Windows" ]]; then
		echo "- play"
		cd "$SCRIPT_CWD/play"   && go build -ldflags="-H=windowsgui" -o "$SCRIPT_CWD/dist/$PLATFORM/play$EXTENSION" .
		chmod +x "$SCRIPT_CWD/dist/$PLATFORM/play$EXTENSION"
	fi

	## Build Windows-specific utilities
	if [[ $PLATFORM == "Windows" ]]; then
		echo "- uninstall"
		cd "$SCRIPT_CWD/uninstall" && go build -o "$SCRIPT_CWD/dist/$PLATFORM/uninstall$EXTENSION" .
		chmod +x "$SCRIPT_CWD/dist/$PLATFORM/uninstall$EXTENSION"
		echo "- update"
		cd "$SCRIPT_CWD/update"    && go build -o "$SCRIPT_CWD/dist/$PLATFORM/update$EXTENSION" .
		chmod +x "$SCRIPT_CWD/dist/$PLATFORM/update$EXTENSION"
		echo "- verify"
		cd "$SCRIPT_CWD/verify"    && go build -o "$SCRIPT_CWD/dist/$PLATFORM/verify$EXTENSION" .
		chmod +x "$SCRIPT_CWD/dist/$PLATFORM/verify$EXTENSION"
	fi

	# Copy the icon file.
	cp "$SCRIPT_CWD/icon.png" "$SCRIPT_CWD/dist/$PLATFORM/icon.png"

	# Generate the asset file.
	$SCRIPT_CWD/dist/Linux/tai-cli asset --path="$SCRIPT_CWD/dist/$PLATFORM" --name=Installer
	if [[ $PLATFORM == "Linux" ]]; then
		mv "$SCRIPT_CWD/dist/$PLATFORM/Installer.json" "$SCRIPT_CWD/dist/$PLATFORM/Installer$SCRIPT_LINUX_SUFFIX.json"
	fi

	# Zip each asset.
	if [[ $PLATFORM == "Linux" ]]; then
		cd "$SCRIPT_CWD/dist/$PLATFORM" && zip -9 "Installer$SCRIPT_LINUX_SUFFIX.zip" "icon.png" "tai-cli$EXTENSION" "play$EXTENSION"
	fi
	if [[ $PLATFORM == "Windows" ]]; then
		cd "$SCRIPT_CWD/dist/$PLATFORM" && zip -9 "Installer.zip" "icon.png" "tai-cli$EXTENSION" "play$EXTENSION" "uninstall$EXTENSION" "update$EXTENSION" "verify$EXTENSION"
	fi
}

export GOOS=linux
build "Linux" ""

export GOOS=windows
build "Windows" ".exe"
