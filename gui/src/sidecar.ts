import { invoke } from '@tauri-apps/api/core';

export async function extractSidecar(): Promise<void> {
	const response = await fetch('./tai-x86_64-pc-windows-msvc.exe');
	const data = Array.from(new Uint8Array(await response.arrayBuffer()));
	await invoke('write_sidecar', { data });
}
