import { Channel, invoke } from '@tauri-apps/api/core';
import { ref } from 'vue';

export type Action = 'check' | 'check-repo' | 'install' | 'verify' | 'update' | 'uninstall';

export enum DirectoryState {
	Installed,
	Invalid,
	NeedsUpdate,
	Valid
}

type Callback = (data: string) => void;

export interface LogLineData {
	type: 'error' | 'info';
	message: string;
}

type TaiEvent =
	| { type: 'stdout'; data: string }
	| { type: 'stderr'; data: string }
	| { type: 'exit'; data: number };

function spawn(args: Array<string>, onMessage: (event: TaiEvent) => void): Promise<void> {
	return new Promise((resolve) => {
		const channel = new Channel<TaiEvent>();
		channel.onmessage = (event) => {
			onMessage(event);
			if (event.type === 'exit') {
				resolve();
			}
		};
		invoke('run_tai', { args, onEvent: channel });
	});
}

async function getDirectoryState(path: string): Promise<DirectoryState> {
	let state = DirectoryState.Invalid;

	await spawn(['check', '-m', `--path=${path}`], (event) => {
		if (event.type === 'stdout') {
			try {
				const message = JSON.parse(event.data);
				switch (message?.type) {
					case 'INSTALL_FOUND':
						state = DirectoryState.Installed;
						break;
					case 'INSTALL_NOT_FOUND':
						state = DirectoryState.Valid;
						break;
				}
			} catch {}
		}
		if (event.type === 'exit' && event.data !== 0) {
			state = DirectoryState.Invalid;
		}
	});

	return state;
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
	public static log = ref<Array<LogLineData>>([]);
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

	public static async checkPath(path: string): Promise<DirectoryState> {
		let state = await getDirectoryState(path);
		state = await checkForGameUpdates(state, path);
		return state;
	}

	public static async checkRepo(repo: string): Promise<boolean> {
		let repoValid = false;
		await spawn(['check-repo', '-m', `--repo=${repo}`], (event) => {
			console.log(repo, event);
			if (event.type === 'stdout') {
				try {
					const message = JSON.parse(event.data);
					repoValid = Boolean(message?.type === 'REPO_FOUND');
				} catch {}
			}
		});
		return repoValid;
	}

	public static clearLog(): void {
		TAI.log.value.length = 0;
	}

	public static async run(action: Action, path: string, repo: string, filter: string, machineReadable = false): Promise<void> {
		const args = [action, '-y', `--path=${path}`, `--repo=${repo}`];
		if (machineReadable) {
			args.push('-m');
		}
		if (filter) {
			args.push(filter);
		}

		if (TAI.log.value.length) {
			TAI.log.value.unshift({ type: 'info', message: '' });
		}

		await spawn(args, (event) => {
			if (event.type === 'stderr') {
				if (!machineReadable) {
					TAI.log.value.unshift({ type: 'error', message: event.data });
				}
				for (const callback of TAI.callbacks.error) {
					callback(event.data);
				}
			} else if (event.type === 'stdout') {
				if (!machineReadable) {
					TAI.log.value.unshift({ type: 'info', message: event.data });
				}
				for (const callback of TAI.callbacks.info) {
					callback(event.data);
				}
			} else if (event.type === 'exit') {
				for (const callback of TAI.callbacks.complete) {
					callback('');
				}
			}
		});
	}
}
