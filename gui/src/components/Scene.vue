<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue';
import { scenes, type StopScene } from '../scenes';

const props = defineProps<{
	scene: keyof typeof scenes;
}>();

const canvas = ref<HTMLCanvasElement | null>(null);
let stopScene: StopScene;

function resetScene (): void {
	stopScene?.();
	if (!canvas.value) return;
	const scene = scenes[props.scene];
	stopScene = scene?.(canvas.value);
}

watch(canvas, resetScene);
watch(() => props.scene, resetScene);

onBeforeUnmount(() => stopScene?.());
</script>

<template>
	<canvas ref="canvas"></canvas>
</template>

<style scoped>
canvas {
	position: fixed;
	inset: 0;
	width: 100vw;
	height: 100vh;
	border-radius: var(--window-border-radius);
}

:root.linux canvas {
	width: calc(100vw - var(--window-shadow) * 2);
	height: calc(100vh - var(--window-shadow) * 2);
}
</style>
