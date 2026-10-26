<script setup lang="ts">
import { WebviewWindow } from '@tauri-apps/api/webviewWindow';
import type { ScrollBarStyle } from '@tauri-apps/api/window';
import { ref } from 'vue';
import { AppWindow } from '../config.ts';
import { loading, store } from '../store.ts';
import { byteCountSI, useFocusRecovery } from '../utilities.ts';
import Icon from './Icon.vue';
import RadioRow from './RadioRow.vue';
import RadioOption from './RadioOption.vue';

const fieldset = ref<HTMLElement>();
const retryButton = ref<HTMLButtonElement>();

// Retrying hides the Retry button, and loading disables the radio buttons.
useFocusRecovery(fieldset, loading.value, () => [
	retryButton.value,
	fieldset.value?.querySelector<HTMLElement>('input:checked'),
]);

async function openAssetSelector(): Promise<void> {
	const existing = await WebviewWindow.getByLabel(AppWindow.AssetSelect);
	if (existing) {
		await existing.setFocus();
		return;
	}

	const webview = new WebviewWindow(AppWindow.AssetSelect, {
		url: 'index.html#/assetSelect',
		title: 'Select Assets to Install',
		width: 450,
		height: 520,
		center: true,
		// WebView2 treats the scrollbar style as an *environment* option, so every
		// webview sharing a data directory has to agree on it. If this differs from
		// `scrollBarStyle` in tauri.conf.json, the window dies as soon as it appears.
		scrollBarStyle: 'fluentOverlay' as ScrollBarStyle.FluentOverlay,
	});
	webview.once('tauri://error', error => console.error('Failed to open the asset selector window:', error));
}
</script>

<template>
	<fieldset ref="fieldset" class="borderless variant-selector">
		<RadioRow
			:disabled="loading.value.value"
			name="variant"
			v-grid:radio
			v-model="store.variant"
		>
			<RadioOption
				image="124.png"
				label="Vanilla"
				value="vanilla"
			>
				<template v-slot:sublabel>
					<span
						class="variant-info"
						:title="`Download size: ${byteCountSI(store.bundleSize.vanilla.packed)}`"
					>
						<span class="visually-hidden">Download size:</span>
						<Icon icon="download" />
						{{ byteCountSI(store.bundleSize.vanilla.packed) }}
					</span>
					<span
						class="variant-info"
						:title="`Installed size: ${byteCountSI(store.bundleSize.vanilla.unpacked)}`"
					>
						<span class="visually-hidden">Installed size:</span>
						<Icon icon="hard-drive" />
						{{ byteCountSI(store.bundleSize.vanilla.unpacked) }}
					</span>
				</template>
			</RadioOption>
			<RadioOption
				image="129.png"
				label="Modded"
				value="modded"
			>
				<template v-slot:sublabel>
					<span
						class="variant-info"
						:title="`Download size: ${byteCountSI(store.bundleSize.modded.packed)}`"
					>
						<span class="visually-hidden">Download size:</span>
						<Icon icon="download" />
						{{ byteCountSI(store.bundleSize.modded.packed) }}
					</span>
					<span
						class="variant-info"
						:title="`Installed size: ${byteCountSI(store.bundleSize.modded.unpacked)}`"
					>
						<span class="visually-hidden">Installed size:</span>
						<Icon icon="hard-drive" />
						{{ byteCountSI(store.bundleSize.modded.unpacked) }}
					</span>
				</template>
			</RadioOption>
			<RadioOption
				image="118.png"
				label="Custom"
				value="custom"
			>
				<template v-slot:sublabel>
					<span
						class="variant-info"
						:title="`Download size: ${byteCountSI(store.bundleSize.custom.packed)}`"
					>
						<span class="visually-hidden">Download size:</span>
						<Icon icon="download" />
						{{ byteCountSI(store.bundleSize.custom.packed) }}
					</span>
					<span
						class="variant-info"
						:title="`Installed size: ${byteCountSI(store.bundleSize.custom.unpacked)}`"
					>
						<span class="visually-hidden">Installed size:</span>
						<Icon icon="hard-drive" />
						{{ byteCountSI(store.bundleSize.custom.unpacked) }}
					</span>
				</template>
			</RadioOption>
		</RadioRow>

		<button
			title="Edit the custom install configuration"
			aria-label="Edit the custom install configuration"
			:class="{ hidden: store.variant !== 'custom' }"
			:inert="store.variant !== 'custom'"
			@click="openAssetSelector()"
		>Edit</button>

		<p
			role="status"
			:class="{
				grey: loading.value.value,
				green: !loading.value.value && store.repositoryFound,
				red: !loading.value.value && !store.repositoryFound,
			}"
		>
			<template v-if="loading.value.value">Checking...</template>
			<template v-else-if="store.repositoryFound">Game files available.</template>
			<template v-else>
				Game files could not be found. Please check your internet connection.
				<button
					ref="retryButton"
					class="retry-button"
					:disabled="loading.value.value"
					@click="store.refreshRepository()"
				>Retry</button>
			</template>
		</p>
	</fieldset>
</template>

<style scoped>
.variant-selector {
	display: grid;
	grid-template:
		'radio radio' auto
		'info  edit ' auto
		/ 2fr  1fr;

	& > button {
		grid-area: edit;
		min-height: 24px;
		font-size: 0.8rem;
		border-radius: 0 0 var(--border-radius) var(--border-radius);
		transition: 0.1s ease opacity;

		&.hidden {
			opacity: 0;
			pointer-events: none;
		}
	}

	& > p {
		grid-area: info;
	}

	& .variant-info:first-of-type {
		margin-inline-end: 2.5ch;
	}

	& .retry-button {
		min-width: unset;
		min-height: unset;
		color: var(--color-text);
	}
}
</style>
