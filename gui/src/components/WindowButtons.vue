<script setup lang="ts">
import { getCurrentWindow } from '@tauri-apps/api/window';
import { openUrl } from '@tauri-apps/plugin-opener';
import { infoUrl } from '../config';

const currentWindow = getCurrentWindow();

async function info(): Promise<void> {
	await openUrl(infoUrl);
}

async function minimize(): Promise<void> {
	await currentWindow.minimize();
}

async function maximize(): Promise<void> {
	await currentWindow.toggleMaximize();
}

async function close(): Promise<void> {
	await currentWindow.close();
}
</script>

<template>
	<div class="window-title-bar">
		<div class="window-drag-region" data-tauri-drag-region></div>
		<div class="window-info">
			<button title="Information" @click="info()"><img src="../assets/info.svg" alt=""></button>
		</div>
		<div class="window-buttons">
			<button class="minimize" title="Minimize" @click="minimize()">—</button>
			<button class="maximize" title="Maximize" @click="maximize()" disabled>▢</button>
			<button class="close" title="Close" @click="close()">✕</button>
		</div>
	</div>
</template>

<style scoped>
.window-title-bar {
	--color-text: #fff;
	--color-text-dim: #4d4d4d;
	--window-button-height: 30px;
	position: fixed;
	top: 0;
	right: 0;
	left: 0;
	display: flex;
	justify-content: space-between;
	height: 70px;
	color: var(--color-text);
	font-family: system-ui, sans-serif;
	user-select: none;
}

.window-drag-region {
	app-region: drag;
	position: absolute;
	inset: 0;
}

.window-info {
	display: flex;
	align-items: center;
	height: var(--window-button-height);
	padding: 3px;

	& button {
		--size: 3ch;
		display: grid;
		place-items: center;
		width: var(--size);
		min-width: var(--size);
		height: var(--size);
		min-height: var(--size);
		padding: 0;
		background: transparent;
		border-radius: 50%;
		opacity: 0.25;
		transition: 0.05s ease opacity;

		&:is(:focus-visible, :hover) {
			opacity: 0.8;
		}
	}
}

.window-buttons {
	--color-close: #c42b1c;
	--color-hover: #2d2d2d;
	display: flex;
	height: max-content;

	& button {
		all: unset;
		box-sizing: border-box;
		display: grid;
		place-items: center;
		width: 45px;
		height: var(--window-button-height);
		color: var(--color-text);
		transition: 0.15s ease background-color;
	}

	& button:disabled {
		color: var(--color-text-dim);
		pointer-events: none;
	}

	& button:not(:disabled):is(:focus-visible, :hover) {
		background: var(--color-hover);
	}

	& .close:not(:disabled):is(:focus-visible, :hover) {
		background: var(--color-close);
	}

	& .minimize {
		padding-block-end: 2px;
	}

	& .maximize {
		padding-block-start: 2px;
	}
}
</style>

<style>
:root.linux .window-title-bar {
	border-radius: calc(var(--window-border-radius) - 1px) calc(var(--window-border-radius) - 1px) 0 0;

	& .window-buttons .close {
		border-radius: 0 calc(var(--window-border-radius) - 1px) 0 0;
	}
}
</style>
