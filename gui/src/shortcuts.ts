import { DEV } from './dev';

export function handleShortcuts(): void {
	window.addEventListener('contextmenu', event => {
		event.preventDefault();
	});

	window.addEventListener('keydown', event => {
		const { ctrlKey: ctrl, key } = event;

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
