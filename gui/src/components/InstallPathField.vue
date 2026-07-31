<script setup lang="ts">
import { open } from '@tauri-apps/plugin-dialog';
import { exists } from '@tauri-apps/plugin-fs';
import { computed, nextTick, watch } from 'vue';
import { DirectoryState, TAI } from '../tai';
import { store } from '../store';
import { loadData } from '../persist';

defineProps<{ disabled: boolean }>();
const noPath = computed(() => !store.path);
const typeImage = computed(() => {
	const state = store.pathState;
	const variant = store.pathVariant;
	if (state === DirectoryState.Invalid) return '80.png';
	switch (variant) {
		case 'vanilla': return '124.png';
		case 'modded': return '129.png';
		case 'custom': return '118.png';
		default: return '80.png';
	}
});

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

		// If an installation exists, check the type.
		if (pathState === DirectoryState.Installed) {
			const data = await loadData();
			store.pathVariant = data?.selectedVariant ?? '';
		} else {
			store.pathVariant = '';
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
watch(() => store.refreshPath, check);
</script>

<template>
	<fieldset class="borderless install-path-field">
		<div
			class="install-path-icon"
			:class="{ invalid: store.pathState === DirectoryState.Invalid }"
		>
			<img
				class="install-path-icon"
				:src="typeImage"
				alt=""
				draggable="false"
			>
		</div>
		<input
			type="text"
			class="install-path-input"
			:disabled
			v-model="store.path"
		>
		<button
			type="button"
			class="primary install-path-button"
			:disabled
			@click="browse()"
		>Browse</button>

		<p class="install-path-info">
			<span v-if="noPath" class="red">No location specified.</span>
			<span v-else-if="store.pathState === DirectoryState.Installed" class="green">Installation found.</span>
			<span v-else-if="store.pathState === DirectoryState.Valid" class="green">Location valid.</span>
			<span v-else class="red">Location invalid.</span>
	
			<template v-if="store.pathState === DirectoryState.Valid">
				<span v-if="store.variant === 'vanilla'">A new vanilla installation will be created here.</span>
				<span v-else-if="store.variant === 'modded'">A new modded installation will be created here.</span>
			</template>
			<template v-else-if="store.pathState === DirectoryState.Installed">
				<span v-if="store.pathVariant === 'vanilla'">A vanilla installation already exists here.</span>
				<span v-else-if="store.pathVariant === 'modded'">A modded installation already exists here.</span>
				<span v-else-if="store.pathVariant === 'custom'">A custom installation already exists here.</span>
			</template>
		</p>
	</fieldset>
</template>

<style scoped>
.install-path-field {
	display: grid;
	grid-template:
		'icon .    input .    button' var(--input-height)
		'.    .    info  info info  ' auto
		/ var(--input-height) calc(var(--gap) * 0) 1fr var(--gap) auto;
	width: 100%;
}

.install-path-icon {
	grid-area: icon;
	display: grid;
	place-items: center;
	padding: 2px;
	background: var(--color-card);
	border-radius: var(--border-radius) 0 0 var(--border-radius);

	&::before {
		content: '';
		position: absolute;
		inset: 0 calc(var(--border-radius) * -1) 0 100%;
		background: inherit;
	}

	& img {
		height: calc(var(--input-height) - 4px);
	}

	&.invalid img {
		filter: grayscale(1);
	}
}

.install-path-input {
	grid-area: input;
	letter-spacing: 0.05ch;
}

.install-path-button {
	grid-area: button;
}

.install-path-info {
	grid-area: info;
}
</style>
