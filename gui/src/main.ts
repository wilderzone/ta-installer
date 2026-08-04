import { WebviewWindow } from '@tauri-apps/api/webviewWindow';
import { getCurrentWindow } from '@tauri-apps/api/window';
import { createApp } from 'vue';
import './root.css';
import App from './App.vue';
import { AppWindow } from './config.ts';
import { vGrid } from './directives.ts';
import { router } from './router.ts';
import { handleShortcuts } from './shortcuts.ts';
import { pinia } from './store.ts';

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
