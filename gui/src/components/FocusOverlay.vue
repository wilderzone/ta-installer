<script setup lang="ts">
import { UnlistenFn } from '@tauri-apps/api/event';
import { getCurrentWindow } from '@tauri-apps/api/window';
import { onMounted, ref } from 'vue';

const overlay = ref<HTMLDivElement>();
let unlistenFocus: UnlistenFn | undefined;

async function handleFocusChange(): Promise<void> {
	unlistenFocus?.();
	const currentWindow = getCurrentWindow();

	unlistenFocus = await currentWindow.onFocusChanged(({ payload: focused }) => {
		if (focused) {
			overlay.value?.classList.remove('unfocused');
		} else {
			overlay.value?.classList.add('unfocused');
		}
	});
}

onMounted(() => {
	void handleFocusChange();
});
</script>

<template>
	<div ref="overlay" class="focus-overlay"></div>
</template>

<style scoped>
.focus-overlay {
	position: fixed;
	inset: 0;
	pointer-events: none;
	transition: 0.1s ease backdrop-filter;

	&.unfocused {
		backdrop-filter: grayscale(0.4);
	}
}
</style>
