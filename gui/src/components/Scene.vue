<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue';
import { scenes, type StopScene } from '../scenes';
import { store } from '../store';

const props = defineProps<{
	scene: keyof typeof scenes;
}>();

const canvas = ref<HTMLCanvasElement | null>(null);
let stopScene: StopScene;

function updateScene (): void {
	stopScene?.();
	if (!canvas.value) return;
	const scene = scenes[props.scene];
	stopScene = scene?.(canvas.value);
}

watch(canvas, updateScene);
watch(() => props.scene, updateScene);
watch(() => store.model, updateScene);

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
}
</style>
