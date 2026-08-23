package main

// Linux has no equivalent of the Windows "Uninstall" registry key: an installed
// application announces itself with a desktop entry instead, which createShortcuts
// already writes. Both of these are no-ops so the commands can call them unconditionally.

func addUninstallerToRegistry(root string) error {
	return nil
}

func removeUninstallerFromRegistry() error {
	return nil
}
