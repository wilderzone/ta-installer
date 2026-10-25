<script setup lang="ts">
import { open } from '@tauri-apps/plugin-dialog';
import { computed } from 'vue';
import { isWindows } from '../config';
import { loading, store } from '../store';
import { DirectoryState } from '../tai';

const noPath = computed(() => !store.path);
const shortcutLabel = isWindows ? 'Create Desktop & Start Menu shortcuts' : 'Create desktop & applications menu shortcuts';
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
	const previous = store.path;
	store.path = await open({
		multiple: false,
		directory: true,
	}) || store.path;

	// Manually refresh the directory if the path was reselected.
	if (previous === store.path) {
		store.refreshDirectory();
	}
}

function change(event: Event): void {
	if (!(event.target instanceof HTMLInputElement)) return;
	store.path = event.target.value;
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
			aria-label="Installation location"
			aria-describedby="install-path-info"
			:aria-invalid="store.step2Error"
			:disabled="loading.value.value"
			:value="store.path"
			@change="change($event)"
		>
		<button
			type="button"
			class="primary install-path-button"
			aria-label="Browse your computer for an installation location"
			:disabled="loading.value.value"
			@click="browse()"
		>Browse</button>

		<label for="create-shortcuts" class="create-shortcuts">
			<input
				type="checkbox"
				id="create-shortcuts"
				name="create-shortcuts"
				v-model="store.createShortcuts"
			>
			<span>{{ shortcutLabel }}</span>
		</label>

		<p id="install-path-info" role="status" class="install-path-info">
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
				<span v-if="store.variant === 'vanilla'">A new vanilla installation will be created here.</span>
				<span v-else-if="store.variant === 'modded'">A new modded installation will be created here.</span>
				<span v-else-if="store.variant === 'custom'">A new custom installation will be created here.</span>
				<span v-else>A new installation will be created here.</span>
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
		'icon     input    .        button  ' var(--input-height)
		'.        info     info     info    ' 1fr
		'.        .        .        .       ' var(--gap-s)
		'shortcut shortcut shortcut shortcut' auto
		/ var(--input-height) 1fr var(--gap-m) auto;
	width: 100%;
	height: 100%;
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

.create-shortcuts {
	grid-area: shortcut;
	display: flex;
	align-items: center;
	gap: var(--fieldset-inline-padding);
	height: var(--input-height);
	padding-inline: var(--fieldset-inline-padding);
	font-size: 0.8rem;
	font-style: italic;
	background: var(--color-card);
	border-radius: var(--border-radius);

	& input[type="checkbox"] {
		margin: 0;

		&:focus-visible {
			outline: var(--focus-outline);
		}
	}
}
</style>
