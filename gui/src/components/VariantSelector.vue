<script setup lang="ts">
import { WebviewWindow } from '@tauri-apps/api/webviewWindow';
import type { ScrollBarStyle } from '@tauri-apps/api/window';
import { nextTick, onMounted, ref } from 'vue';
import { repository } from '../config.ts';
import { store } from '../store.ts';
import { TAI } from '../tai';
import RadioRow from './RadioRow.vue';
import RadioOption from './RadioOption.vue';

defineProps<{ disabled: boolean }>();
const model = defineModel<string>({ required: true });
const repositoryFound = ref(false);

async function delay(duration: number): Promise<void> {
	return new Promise(resolve => {
		window.setTimeout(resolve, duration);
	});
}

async function openAssetSelector(): Promise<void> {
	const windowLabel = 'assetSelect';
	const existing = await WebviewWindow.getByLabel(windowLabel);
	if (existing) {
		await existing.setFocus();
		return;
	}

	const webview = new WebviewWindow(windowLabel, {
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

async function check(): Promise<void> {
	nextTick(async () => {
		store.loading = true;
		const found = await TAI.checkRepo(repository);
		const manifest = await TAI.getRemoteManifest(repository);
		store.requiredCustomAssets = new Set(TAI.determineRequiredAssets(manifest?.assets ?? []).map(asset => asset.name));
		await delay(500);
		store.step1Error = !found;
		repositoryFound.value = found;
		store.loading = false;
	});
}

onMounted(() => {
	void check();
});
</script>

<template>
	<fieldset class="borderless">
		<RadioRow
			:disabled
			name="variant"
			v-model="model"
		>
			<RadioOption
				image="124.png"
				label="Vanilla"
				sublabel="(4.5GB)"
				value="vanilla"
			/>
			<RadioOption
				image="129.png"
				label="Modded"
				sublabel="(8GB)"
				value="modded"
			/>
			<RadioOption
				image="118.png"
				label="Custom"
				value="custom"
				@click="openAssetSelector()"
			/>
		</RadioRow>
		<p v-if="store.loading" class="grey">Checking...</p>
		<p v-else-if="repositoryFound" class="green">Game files available.</p>
		<p v-else-if="!repositoryFound" class="red">Game files could not be found. Please check your internet connection.</p>
	</fieldset>
</template>

<style scoped>
</style>
