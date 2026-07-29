<script setup lang="ts">
import { open } from '@tauri-apps/plugin-dialog';
import { exists } from '@tauri-apps/plugin-fs';
import { computed, nextTick, watch } from 'vue';
import { DirectoryState, TAI } from '../tai';
import { store } from '../store';

const props = defineProps<{ disabled: boolean }>();
const noPath = computed(() => !store.path);

async function browse(): Promise<void> {
	store.path = await open({
		multiple: false,
		directory: true,
	}) || store.path;

	await check();
}

async function check(): Promise<void> {
	nextTick(async () => {
		const path = store.path;
		const parent = parentDirectory(path);

		// Reject empty paths.
		if (!path) {
			store.pathState = DirectoryState.Invalid;
			return;
		}

		// Check that the chosen directory and its parent exist.
		const pathExists = await exists(path);
		const parentExists = await exists(parent);
		if (!(pathExists || pathEndsWithTA(path)) || !parentExists) {
			store.pathState = DirectoryState.Invalid;
			return;
		}

		// Check the normalised path with TAI.
		const normalisedPath = normalisePath(path);
		let pathState = await TAI.checkPath(normalisedPath);

		// If the normalised path returns valid but not installed,
		// check the actual chosen path for an installation instead.
		if (pathState === DirectoryState.Valid) {
			pathState = await TAI.checkPath(path);
		}

		store.pathState = pathState;
	});
}

function pathEndsWithTA(path: string): boolean {
	return /[\\\/]Tribes Ascend[\\\/]?$/.test(path);
}

function normalisePath(path: string): string {
	if (pathEndsWithTA(path)) return path;
	else return `${path}\\Tribes Ascend`.replaceAll(/[\\\/]+/g, '\\');
}

function parentDirectory(path: string): string {
	const segments = path.replace(/[\\\/]$/, '').split(/[\\\/]/);
	segments.pop();
	return segments.join('\\');
}

watch(() => store.path, check, { immediate: true });
</script>

<template>
	<fieldset class="borderless">
		<div class="row">
			<input
				type="text"
				class="install-path-input"
				:disabled
				v-model="store.path"
			>
			<button
				type="button"
				class="primary"
				:disabled
				@click="browse()"
			>Browse</button>
		</div>
		<p v-if="noPath" class="red">No location specified.</p>
		<p v-else-if="store.pathState === DirectoryState.Installed" class="green">Installation found.</p>
		<p v-else-if="store.pathState === DirectoryState.Valid" class="green">Location valid.</p>
		<p v-else class="red">Location invalid.</p>
	</fieldset>
</template>

<style scoped>
.install-path-input {
	letter-spacing: 0.05ch;
}
</style>
