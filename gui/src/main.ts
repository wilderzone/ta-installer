import { invoke } from '@tauri-apps/api/core';
import { sep } from '@tauri-apps/api/path';
import { WebviewWindow } from '@tauri-apps/api/webviewWindow';
import { getCurrentWindow } from '@tauri-apps/api/window';
import { createApp } from 'vue';
import './root.css';
import App from './App.vue';
import { AppWindow, useDefaultInstallPath, usePlatform } from './config.ts';
import { vGrid } from './directives.ts';
import { router } from './router.ts';
import { handleShortcuts } from './shortcuts.ts';
import { pinia, store } from './store.ts';
import { usePathSeparator } from './utilities.ts';

/**
 * Learn what platform this is, then start the app.
 *
 * The separator is settled synchronously, before anything can await, because the store's
 * path watcher fires on a microtask and the path helpers it reaches for are synchronous.
 * The default install path has to be asked for, so the store starts out empty and is filled
 * in here - nothing is rendered until the mount below, so the empty value is never seen.
 */
async function start(): Promise<void> {
	const separator = sep();
	usePathSeparator(separator);
	usePlatform(separator);

	const path = await invoke<string>('default_install_path');
	useDefaultInstallPath(path);
	store.path = path;

	createApp(App)
		.use(router)
		.use(pinia)
		.directive('grid', vGrid)
		.mount('#app');

	handleShortcuts();

	// Close auxiliary windows when the main window is closed.
	const currentWindow = getCurrentWindow();
	currentWindow.onCloseRequested(async () => {
		if (currentWindow.label === AppWindow.Main) {
			const assetSelectWindow = await WebviewWindow.getByLabel(AppWindow.AssetSelect);
			assetSelectWindow?.close();
		}
	});
}

start();
