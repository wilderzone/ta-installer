import { DEV } from './dev';

export function cdn(path: string) {
	const base = DEV ? 'http://localhost:3000/' : 'https://cdn.wilderzone.org/';
	return new URL(path, base);
}

export enum AppWindow {
	Main = 'main',
	AssetSelect = 'assetSelect'
}

export const infoUrl = 'https://wilderzone.org/docs/ta-installer';

/**
 * Where the game is installed by default.
 * The value is platform-specific, so it comes from the Rust backend during startup rather than
 * being hardcoded here.
 */
export let defaultInstallPath = '';
export function useDefaultInstallPath(value: string): void {
	defaultInstallPath = value;
}

/**
 * Whether the app is running on Windows. Set during startup, alongside the path default.
 * The path separator is the tell, which avoids pulling in a plugin just to check the platform.
 */
export let isWindows = true;
export function usePlatform(separator: string): void {
	isWindows = separator === '\\';
}
export const models = ['GravCycle', 'GravCycle', 'Shrike', 'Beowulf'];
export const repository = cdn('ta/game').href;
export const assetInclusions: Record<string, Array<string>> = {
	vanilla: ['core','Air Arena','Arx Novena','Bella Omega','Blueshift','Canyon Crusade Revival','Crossfire','Dangerous Crossing','Drydock','Fraytown','Hellfire','Hinterland','Ice Coaster','Katabatic','Lavarena','Perdition','Permafrost','Raindance','Stonehenge','Tartarus','Temple Ruins','Terminus','Undercroft','Walledin','Whiteout'],
	modded: [],
	custom: []
};
