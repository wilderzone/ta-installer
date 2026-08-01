<script setup lang="ts">
import { getCurrentWindow } from '@tauri-apps/api/window';

const currentWindow = getCurrentWindow();

async function close(): Promise<void> {
	await currentWindow.close();
}
</script>

<template>
	<div class="asset-navigation">
		<button class="primary" @click="close()">OK</button>
	</div>
</template>

<style scoped>
.asset-navigation {
	display: flex;
	justify-content: end;
	align-items: center;
	gap: var(--gap-m);
	padding: var(--gap-s);
}
</style>
