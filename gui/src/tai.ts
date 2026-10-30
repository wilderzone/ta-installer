import { Channel, invoke } from '@tauri-apps/api/core';
import { ref } from 'vue';
import { Asset, BundleSize, Manifest } from './manifest';

export type Action = 'check' | 'check-repo' | 'install' | 'verify' | 'update' | 'uninstall';

export const verbs: Record<Action, string> = {
	'check':      'Check',
	'check-repo': 'Check',
	'install':    'Installation',
	'verify':     'Verification',
	'update':     'Update',
	'uninstall':  'Uninstall'
};

export enum DirectoryState {
	Installed,
	Invalid,
	NeedsUpdate,
	Valid
}

type Callback = (data: string) => void;

export interface DownloadProgress {
	/** The bytes received so far. */
	bytes: number;
	/** The bytes expected in total. */
	size: number;
	/** The transfer rate, in bytes per second. */
	rate: number;
}

export interface LogLineData {
	type: 'error' | 'info';
	message: string;
}

type TaiEvent =
	| { type: 'stdout'; data: string }
	| { type: 'stderr'; data: string }
	| { type: 'exit'; data: number };

/**
 * Run the CLI, resolving once it exits.
 * A cancellable run registers itself as the one `TAI.cancel` stops, so the short-lived
 * probes below can never be mistaken for the action the user is waiting on.
 */
function spawn(args: Array<string>, onMessage: (event: TaiEvent) => void, cancellable = false): Promise<void> {
	return new Promise((resolve) => {
		const channel = new Channel<TaiEvent>();
		channel.onmessage = (event) => {
			onMessage(event);
			if (event.type === 'exit') {
				resolve();
			}
		};
		invoke('run_tai', { args, cancellable, onEvent: channel });
	});
}

export interface PathCheck {
	state: DirectoryState;
	/** The bytes free on the path's drive, or `undefined` if free space couldn't be measured. */
	freeSpace: number | undefined;
}

async function checkDirectory(path: string): Promise<PathCheck> {
	const check: PathCheck = { state: DirectoryState.Invalid, freeSpace: undefined };

	await spawn(['check', '-m', `--path=${path}`], (event) => {
		if (event.type === 'stdout') {
			try {
				const message = JSON.parse(event.data);
				switch (message?.type) {
					case 'FREE_SPACE':
						check.freeSpace = Number(message.bytes ?? 0);
						break;
					case 'INSTALL_FOUND':
						check.state = DirectoryState.Installed;
						break;
					case 'INSTALL_NOT_FOUND':
						check.state = DirectoryState.Valid;
						break;
				}
			} catch {}
		}
		if (event.type === 'exit' && event.data !== 0) {
			check.state = DirectoryState.Invalid;
		}
	});

	return check;
}

async function checkForGameUpdates(state: DirectoryState, path: string): Promise<DirectoryState> {
	if (state !== DirectoryState.Installed) {
		return state;
	}

	await spawn(['check-updates', '-m', `--path=${path}`], (event) => {
		if (event.type === 'stdout') {
			try {
				const message = JSON.parse(event.data);
				switch (message?.type) {
					case 'UPDATE_FOUND':
						state = DirectoryState.NeedsUpdate;
						break;
					case 'INSTALL_UP_TO_DATE':
						state = DirectoryState.Installed;
						break;
				}
			} catch {}
		}
	});

	return state;
}

export class TAI {
	private static callbacks = {
		info: <Array<Callback>> [],
		error: <Array<Callback>> [],
		complete: <Array<Callback>> []
	};
	/** The download the CLI is currently measuring, or `undefined` if there isn't one. */
	public static download = ref<DownloadProgress | undefined>();
	public static driversHint = ref<string | undefined>();
	public static log = ref<Array<LogLineData>>([]);
	/** The CLI's warning that the drive may not have room for the downloads, if it gave one during this run. */
	public static shortOfSpace = ref<string | undefined>();
	public static status = ref();

	public static attach(onInfo: Callback, onError: Callback, onComplete: Callback): void {
		TAI.on('info', onInfo);
		TAI.on('error', onError);
		TAI.on('complete', onComplete);
	}

	public static on(event: 'error' | 'info' | 'complete', callback: Callback): void {
		TAI.callbacks[event].push(callback);
	}

	public static off(event: 'error' | 'info' | 'complete', callback: Callback): void {
		const index = TAI.callbacks[event].findIndex(c => c === callback);
		TAI.callbacks[event].splice(index, 1);
	}

	/**
	 * Ask the running action to stop.
	 * The CLI stops at its next safe point, logs what was interrupted and how to resume,
	 * then exits normally - so the in-flight `TAI.run` still resolves through its own
	 * `complete` callbacks. This method is safe to call multiple times.
	 */
	public static cancel(): Promise<void> {
		return invoke('cancel_tai');
	}

	public static async checkPath(path: string): Promise<PathCheck> {
		const check = await checkDirectory(path);
		check.state = await checkForGameUpdates(check.state, path);
		return check;
	}

	/** Measure the free space on the path's drive, without checking for updates. */
	public static async freeSpace(path: string): Promise<number | undefined> {
		return (await checkDirectory(path)).freeSpace;
	}

	public static async checkRepo(repo: string): Promise<boolean> {
		let repoValid = false;
		await spawn(['check-repo', '-m', `--repo=${repo}`], (event) => {
			if (event.type === 'stdout') {
				try {
					const message = JSON.parse(event.data);
					repoValid = Boolean(message?.type === 'REPO_FOUND');
				} catch {}
			}
		});
		return repoValid;
	}

	public static determineRequiredAssets(assets: Array<Asset>): Array<Asset> {
		const output = new Set<Asset>();
		for (const asset of assets) {
			if (asset.type === 'core') {
				output.add(asset);
			}
			if (!asset.dependencies.length) {
				continue;
			}
			for (const dependency of asset.dependencies) {
				const dep = assets.find(a => a.type === dependency || a.name === dependency);
				if (dep) {
					output.add(dep);
				}
			}
		}
		return Array.from(output);
	}

	public static determineBundleSize(assets: Array<Asset>): BundleSize {
		const packed = assets.reduce((sum, asset) => sum + asset.packedSize, 0);
		const unpacked = assets.reduce((sum, asset) => sum + asset.unpackedSize, 0);
		return { packed, unpacked };
	}

	/**
	 * Build a filter flag for the CLI.
	 * Names are never quoted: the flag is passed as a single argv element (see `run_tai`),
	 * so it never touches a shell and spaces need no protection. Quoting it would make the
	 * quotes part of the asset name, and the CLI would silently ignore the asset.
	 */
	public static filter(type: 'include' | 'exclude' | 'remove' | 'keep', items: Array<string>): string {
		return `--${type}=${items.join(',')}`;
	}

	public static async getLocalManifest(path: string): Promise<Manifest | undefined> {
		let manifest: Manifest | undefined;
		await spawn(['list', '-m', `--path=${path}`], (event) => {
			if (event.type === 'stdout') {
				try {
					const message = JSON.parse(event.data);
					if (message?.type === 'MANIFEST_CONTENTS') {
						manifest = JSON.parse(message.message);
					}
				} catch {}
			}
		});
		return manifest;
	}

	public static async getRemoteManifest(repo: string): Promise<Manifest | undefined> {
		let manifest: Manifest | undefined;
		await spawn(['list', '-m', `--repo=${repo}`], (event) => {
			if (event.type === 'stdout') {
				try {
					const message = JSON.parse(event.data);
					if (message?.type === 'MANIFEST_CONTENTS') {
						manifest = JSON.parse(message.message);
					}
				} catch {}
			}
		});
		return manifest;
	}

	public static clearLog(): void {
		TAI.log.value.length = 0;
	}

	/**
	 * Interpret a line as one of the structured events the CLI emits under `-p`.
	 * @returns `true` if the line was an event, `false` if it was an ordinary log message.
	 */
	private static intercept(data: string): boolean {
		if (!data.startsWith('{')) {
			return false;
		}
		try {
			const message = JSON.parse(data);
			switch (message?.type) {
				case 'DOWNLOAD_STARTED':
				case 'DOWNLOAD_PROGRESS':
					TAI.download.value = {
						bytes: Number(message.bytes ?? 0),
						size: Number(message.size ?? 0),
						rate: Number(message.rate ?? 0)
					};
					return true;
				case 'DOWNLOAD_COMPLETE':
					TAI.download.value = undefined;
					return true;
				case 'RUNTIME_DRIVERS_HINT':
					TAI.driversHint.value = String(message.message ?? '');
					return true;
				case 'INSUFFICIENT_SPACE':
					TAI.shortOfSpace.value = String(message.message ?? '');
					return true;
			}
		} catch {}
		return false;
	}

	/**
	 * Run an action, resolving once it exits.
	 * @param stopIfShort Stop before downloading anything if there may not be enough free space,
	 * rather than going ahead. {@link TAI.shortOfSpace} then holds the CLI's warning.
	 */
	public static async run(action: Action, path: string, repo: string, filter: string, shortcuts = true, stopIfShort = false): Promise<void> {
		const args = [action, '-y', '-p', `--path=${path}`, `--repo=${repo}`];
		if (!shortcuts) {
			args.push('--shortcuts=false');
		}
		if (stopIfShort) {
			args.push('--stop-if-short');
		}
		if (filter) {
			args.push(filter);
		}

		if (TAI.log.value.length) {
			TAI.log.value.unshift({ type: 'info', message: '' });
		}

		TAI.download.value = undefined;
		TAI.driversHint.value = undefined;
		TAI.shortOfSpace.value = undefined;
		await spawn(args, (event) => {
			if (event.type === 'stderr') {
				TAI.log.value.unshift({ type: 'error', message: event.data });
				for (const callback of TAI.callbacks.error) {
					callback(event.data);
				}
			} else if (event.type === 'stdout') {
				if (TAI.intercept(event.data)) {
					return;
				}
				TAI.log.value.unshift({ type: 'info', message: event.data });
				for (const callback of TAI.callbacks.info) {
					callback(event.data);
				}
			} else if (event.type === 'exit') {
				TAI.download.value = undefined;
				for (const callback of TAI.callbacks.complete) {
					callback('');
				}
			}
		}, true);
	}
}
