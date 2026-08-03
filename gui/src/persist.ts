import { mkdir, readTextFile, writeTextFile } from '@tauri-apps/plugin-fs';
import { store } from './store';
import { join } from '@tauri-apps/api/path';

const filename = 'tai-gui.json';
const directory = 'Installer';

interface TAIGUIData {
	selectedVariant: string;
}

export async function saveData(): Promise<void> {
	const data: TAIGUIData = {
		selectedVariant: store.variant
	};
	try {
		// TAI creates this directory when it downloads the installation utilities, which an
		// interrupted install may not have reached yet. Attempt to create it just in case.
		const path = await join(store.path, directory);
		await mkdir(path, { recursive: true });
		await writeTextFile(await join(path, filename), JSON.stringify(data));
	} catch (error) {
		// This data file is non-essential, so we can dispose of any errors.
		console.error(error);
	}
}

export async function loadData(): Promise<TAIGUIData | undefined> {
	try {
		const path = await join(store.path, directory, filename);
		const text = await readTextFile(path);
		const data = JSON.parse(text);
		if (!data) return undefined;
		const safe: TAIGUIData = {
			selectedVariant: parseVariant(data.selectedVariant)
		};
		return safe;
	} catch (error) {
		// This data file is non-essential, so we can dispose of any errors.
		console.error(error);
		return undefined;
	}
}

function parseVariant(variant: string | undefined): string {
	switch (variant) {
		case 'vanilla': return 'vanilla';
		case 'modded': return 'modded';
		case 'custom': return 'custom';
		default: return 'custom';
	}
}
