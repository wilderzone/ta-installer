<script setup lang="ts">
import { writeText } from '@tauri-apps/plugin-clipboard-manager';
import { ref, watch } from 'vue';
import { isWindows } from '../config';
import { TAI } from '../tai';

const open = defineModel<boolean>('open', { required: true });
const dialog = ref<HTMLDialogElement>();
const copied = ref(false);
const elevatedTerm = isWindows ? 'administrator' : 'elevated';

watch(TAI.driversHint, value => {
	if (value) {
		copied.value = false;
		open.value = true;
	}
});

// Not `showModal()`, because a modal dialog covers the window buttons and makes them inert.
// The view makes the rest of the window inert instead. This runs after the DOM updates, so
// the window is interactive again by the time closing the dialog restores focus to it.
watch(open, value => {
	if (value) dialog.value?.show();
	else dialog.value?.close();
}, { flush: 'post' });

function copy(): void {
	if (!TAI.driversHint.value) {
		return;
	}
	void writeText(TAI.driversHint.value);
	copied.value = true;
}
</script>

<template>
	<div v-show="open" class="drivers-dialog">
		<dialog
			ref="dialog"
			class="panel"
			aria-labelledby="drivers-dialog-title"
			@keydown.esc="open = false"
		>
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
		</dialog>
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
		flex-direction: column;
		gap: var(--gap-s);
		max-width: 62ch;
		margin: 0;
		padding: var(--gap-m);
		color: var(--color-text);
		background: var(--color-background);
		border: 1px solid var(--color-border);
		border-radius: var(--border-radius);
		box-shadow: 0 8px 24px #0008;

		/* Only while open, so the closed dialog stays hidden. */
		&[open] {
			display: flex;
		}
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
