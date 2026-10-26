import { openUrl } from '@tauri-apps/plugin-opener';
import { infoUrl } from './config';
import { DEV } from './dev';

export function handleShortcuts(): void {
	window.addEventListener('contextmenu', event => {
		event.preventDefault();
	});

	window.addEventListener('keydown', event => {
		const { ctrlKey: ctrl, key } = event;

		if (key === 'F1') {
			event.preventDefault();
			void openUrl(infoUrl);
			return;
		}

		if (key === 'F3') {
			event.preventDefault();
			return;
		}

		if (key === 'F5' && !DEV) {
			event.preventDefault();
			return;
		}

		if (key === 'F7') {
			event.preventDefault();
			return;
		}

		if (key === 'F12' && !DEV) {
			event.preventDefault();
			return;
		}

		if (ctrl && key === 'f') {
			event.preventDefault();
			return;
		}

		if (ctrl && key === 'g') {
			event.preventDefault();
			return;
		}

		if (ctrl && key === 'j') {
			event.preventDefault();
			return;
		}

		if (ctrl && key === 'p') {
			event.preventDefault();
			return;
		}

		if (ctrl && key === 'r') {
			event.preventDefault();
			return;
		}
	});
}
