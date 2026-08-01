import { DEV } from './dev';

export function cdn(path: string) {
	return new URL(path, 'https://cdn.wilderzone.live/');
}

export const customInstructions = 'https://github.com/wilderzone/ta-installer#cli';
export const defaultInstallPath = 'C:\\ProgramData\\Tribes Ascend';
export const models = ['GravCycle', 'GravCycle', 'Shrike', 'Beowulf'];
export const repository = DEV ? 'http://localhost:3000/ta' : 'https://cdn.wilderzone.live/ta';
export const assetInclusions: Record<string, Array<string>> = {
	vanilla: ['core','Air Arena','Arx Novena','Bella Omega','Blueshift','Canyon Crusade Revival','Crossfire','Dangerous Crossing','Drydock','Fraytown','Hellfire','Hinterland','Ice Coaster','Katabatic','Lavarena','Perdition','Permafrost','Raindance','Stonehenge','Tartarus','Temple Ruins','Terminus','Undercroft','Walledin','Whiteout'],
	modded: [],
	custom: []
};
