<script setup lang="ts">
import { open } from '@tauri-apps/plugin-dialog';
import { computed } from 'vue';
import { loading, store } from '../store';
import { DirectoryState } from '../tai';

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
}
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
			:disabled="loading.value.value"
			v-model="store.path"
		>
		<button
			type="button"
			class="primary install-path-button"
			:disabled="loading.value.value"
			@click="browse()"
		>Browse</button>

		<p class="install-path-info">
			<span v-if="noPath" class="red">No location specified.</span>
			<span v-else-if="store.pathState === DirectoryState.Installed || store.pathState === DirectoryState.NeedsUpdate" class="green">
				<template v-if="store.pathVariant === 'vanilla'">Vanilla installation found.</template>
				<template v-else-if="store.pathVariant === 'modded'">Modded installation found.</template>
				<template v-else-if="store.pathVariant === 'custom'">Custom installation found.</template>
				<template v-else>Installation found.</template>
			</span>
			<span v-else-if="store.pathState === DirectoryState.Valid" class="green">Location valid.</span>
			<span v-else class="red">Location invalid.</span>
	
			<template v-if="store.pathState === DirectoryState.Valid">
				<span v-if="store.variant === 'vanilla'"><br>A new vanilla installation will be created here.</span>
				<span v-else-if="store.variant === 'modded'"><br>A new modded installation will be created here.</span>
				<span v-else-if="store.variant === 'custom'"><br>A new custom installation will be created here.</span>
			</template>

			<template v-if="store.pathState === DirectoryState.NeedsUpdate">
				<span class="yellow"><br>A game update is available.</span>
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
		/ var(--input-height) 0 1fr var(--gap-m) auto;
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
