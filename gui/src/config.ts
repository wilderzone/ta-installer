import { DEV } from './dev';

export const repository: Record<string, string> = {
	vanilla: DEV ? 'http://localhost:3000/ta' : 'https://cdn.wilderzone.live/ta',
	modded: DEV ? 'http://localhost:3000/ta-modded' : 'https://cdn.wilderzone.live/ta-modded',
	custom: 'https://github.com/wilderzone/ta-installer#cli'
};
