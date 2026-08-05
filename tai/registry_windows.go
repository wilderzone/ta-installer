package main

import (
	"path/filepath"
	"time"

	"golang.org/x/sys/windows/registry"
)

// addUninstallerToRegistry creates a registry entry for TA under the Windows "Uninstall" key.
// This allows the game to be uninstalled by Windows (eg, via the Control Panel).
// This function should be called whenever the game is updated/modified to keep the "InstallDate" accurate.
// See: https://learn.microsoft.com/en-gb/windows/win32/msi/uninstall-registry-key
func addUninstallerToRegistry(root string) error {
	rootKey := registry.CURRENT_USER

	key, _, err := registry.CreateKey(rootKey, UNINSTALLER_REGISTRY_PATH, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer key.Close()

	executable := filepath.Join(root, "Binaries", "Win32", "TribesAscend.exe")
	err = key.SetStringValue("DisplayIcon", executable)
	if err != nil {
		return err
	}

	err = key.SetStringValue("DisplayName", PRODUCT_NAME)
	if err != nil {
		return err
	}

	err = key.SetStringValue("DisplayVersion", UNINSTALLER_DISPLAY_VERSION)
	if err != nil {
		return err
	}

	bytes, err := directorySize(root)
	// The "EstimatedSize" field is not crucial, so if the directory size check fails just skip this value.
	if err == nil {
		size := uint32(bytes / 1024) // Convert bytes to KB.
		err = key.SetDWordValue("EstimatedSize", size)
		if err != nil {
			return err
		}
	}

	err = key.SetStringValue("HelpLink", UNINSTALLER_HELP_LINK)
	if err != nil {
		return err
	}

	// The install date should be updated whenever the game installation
	// is updated/modified, not just when initially installed.
	date := time.Now().Format("20060102")
	err = key.SetStringValue("InstallDate", date)
	if err != nil {
		return err
	}

	err = key.SetStringValue("InstallLocation", root)
	if err != nil {
		return err
	}

	err = key.SetStringValue("Publisher", UNINSTALLER_PUBLISHER)
	if err != nil {
		return err
	}

	// Both strings are command lines, not paths, so the executable has to be quoted:
	// an unquoted install path containing a space would otherwise be read as a command
	// plus arguments.
	uninstallString := `"` + filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY, "uninstall.exe") + `"`
	quietUninstallString := uninstallString + " -s"
	err = key.SetStringValue("UninstallString", uninstallString)
	if err != nil {
		return err
	}
	err = key.SetStringValue("QuietUninstallString", quietUninstallString)
	if err != nil {
		return err
	}

	err = key.SetStringValue("URLInfoAbout", UNINSTALLER_ABOUT_LINK)
	if err != nil {
		return err
	}

	err = key.SetStringValue("URLUpdateInfo", UNINSTALLER_UPDATES_LINK)
	if err != nil {
		return err
	}

	return nil
}

// removeUninstallerFromRegistry removes the registry entry for TA from the Windows "Uninstall" key.
func removeUninstallerFromRegistry() error {
	rootKey := registry.CURRENT_USER

	err := registry.DeleteKey(rootKey, UNINSTALLER_REGISTRY_PATH)
	if err != nil {
		return err
	}

	return nil
}
