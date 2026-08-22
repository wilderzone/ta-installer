<script setup lang="ts">
import { writeText } from '@tauri-apps/plugin-clipboard-manager';
import { ref, watch } from 'vue';
import { isWindows } from '../config';
import { TAI } from '../tai';

const open = ref(false);
const copied = ref(false);
const elevatedTerm = isWindows ? 'administrator' : 'elevated';

watch(TAI.driversHint, value => {
	if (value) {
		copied.value = false;
		open.value = true;
	}
});

function copy(): void {
	if (!TAI.driversHint.value) {
		return;
	}
	void writeText(TAI.driversHint.value);
	copied.value = true;
}
</script>

<template>
	<div v-if="open" class="drivers-dialog" role="dialog" aria-modal="true" aria-labelledby="drivers-dialog-title">
		<div class="panel">
			<h2 id="drivers-dialog-title"><span>⚠ 32-bit Graphics Drivers are Missing</span></h2>
			<p>
				Tribes: Ascend is a 32-bit program, requiring 32-bit drivers for your graphics card.
				The game is installed and ready, but it won't start until these drivers are present.
			</p>
			<p>
				Because installing these drivers requires {{ elevatedTerm }} permissions, TAI is unable to run these commands for you.
			</p>
			<p>Run this in a terminal, then start the game:</p>
			<code>{{ TAI.driversHint.value }}</code>
			<div class="actions">
				<button @click="copy()">{{ copied ? 'Copied' : 'Copy' }}</button>
				<button class="primary" @click="open = false">OK</button>
			</div>
		</div>
	</div>
</template>

<style scoped>
.drivers-dialog {
	position: fixed;
	inset: 0;
	display: grid;
	place-items: center;
	padding: var(--gap-m);
	background: #0009;
	backdrop-filter: blur(4px);

	& .panel {
		display: flex;
		flex-direction: column;
		gap: var(--gap-s);
		max-width: 62ch;
		padding: var(--gap-m);
		background: var(--color-background);
		border: 1px solid var(--color-border);
		border-radius: var(--border-radius);
		box-shadow: 0 8px 24px #0008;
	}

	& h2 {
		display: flex;
		margin: calc(var(--gap-m) * -1) calc(var(--gap-m) * -1) var(--gap-xs);

		& span {
			flex-shrink: 0;
			padding: var(--gap-xs) var(--gap-m);
			padding-inline-end: calc(var(--gap-m) * 2);
			color: var(--color-text-yellow);
			text-box: trim-start;
			background: var(--color-card);
			clip-path: polygon(
				0% 0%,
				100% 0%,
				calc(100% - var(--gap-m)) 100%,
				0% 100%
			);
		}
	}

	& p {
		margin: 0;
		color: var(--color-text);
	}

	& code {
		padding: var(--gap-s);
		background: var(--color-input);
		border-radius: var(--border-radius);
		overflow-wrap: break-word;
		user-select: all;
	}

	& .actions {
		display: flex;
		justify-content: end;
		gap: var(--gap-s);
		margin-top: var(--gap-xs);
	}
}
</style>
