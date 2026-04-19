import { Command } from '@tauri-apps/plugin-shell';
import { onMounted, onUnmounted, ref } from 'vue';

export type Action = 'check' | 'check-repo' | 'install' | 'verify' | 'update' | 'uninstall';

type Callback = (data: string) => void;

export interface LogLineData {
	type: 'error' | 'info';
	message: string;
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
		onMounted(() => {
			TAI.on('info', onInfo);
			TAI.on('error', onError);
			TAI.on('complete', onComplete);
		});
		onUnmounted(() => {
			TAI.off('info', onInfo);
			TAI.off('error', onError);
			TAI.off('complete', onComplete);
		});
	}

	public static on(event: 'error' | 'info' | 'complete', callback: Callback): void {
		TAI.callbacks[event].push(callback);
	}

	public static off(event: 'error' | 'info' | 'complete', callback: Callback): void {
		const index = TAI.callbacks[event].findIndex(c => c === callback);
		TAI.callbacks[event].splice(index, 1);
	}

	public static async checkPath(path: string): Promise<boolean> {
		return new Promise(async (resolve) => {
			let pathValid = false;
			const args = ['check', '-m', `--path=${path}`];
			const command = Command.sidecar('binaries/tai', args);
			command.stderr.on('data', () => {
				pathValid = false;
			});
			command.stdout.on('data', (data) => {
				const message = JSON.parse(data);
				pathValid = Boolean(message?.type === 'INSTALL_FOUND');
			});
			command.on('close', () => {
				resolve(pathValid);
			});
			command.on('error', () => {
				resolve(pathValid);
			});
			await command.spawn();
		});
	}

	public static async checkRepo(repo: string): Promise<boolean> {
		return new Promise(async (resolve) => {
			let repoValid = false;
			const args = ['check-repo', '-m', `--repo=${repo}`];
			const command = Command.sidecar('binaries/tai', args);
			command.stderr.on('data', () => {
				repoValid = false;
			});
			command.stdout.on('data', (data) => {
				const message = JSON.parse(data);
				repoValid = Boolean(message?.type === 'REPO_FOUND');
			});
			command.on('close', () => {
				resolve(repoValid);
			});
			command.on('error', () => {
				resolve(repoValid);
			});
			await command.spawn();
		});
	}

	public static clearLog(): void {
		TAI.log.value.length = 0;
	}

	public static async run(action: Action, path: string, repo: string, machineReadable = false): Promise<void> {
		const args = [action, '-y', `--path=${path}`, `--repo=${repo}`];
		if (machineReadable) {
			args.push('-m');
		}

		const command = Command.sidecar('binaries/tai', args);
		command.stderr.on('data', (data) => {
			if (!machineReadable) {
				TAI.log.value.unshift({ type: 'error', message: data });
			}
			for (const callback of TAI.callbacks.error) {
				callback(data);
			}
		});
		command.stdout.on('data', (data) => {
			if (!machineReadable) {
				TAI.log.value.unshift({ type: 'info', message: data });
			}
			for (const callback of TAI.callbacks.info) {
				callback(data);
			}
		});
		command.on('close', () => {
			for (const callback of TAI.callbacks.complete) {
				callback('');
			}
		});
		await command.spawn();
	}
}
