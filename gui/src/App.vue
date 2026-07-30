<script setup lang="ts">
import Background from './components/Background.vue';
import Card from './components/Card.vue';
import Configurator from './components/Configurator.vue';
import FocusOverlay from './components/FocusOverlay.vue';
import Logo from './components/Logo.vue';
import Scene from './components/Scene.vue';
import WindowButtons from './components/WindowButtons.vue';
</script>

<template>
	<Background />
	<Scene scene="home" />
	<main>
		<Logo v-grid:logo />
		<Card v-grid:steps name="Installation configuration">
			<Configurator />
		</Card>
	</main>
	<FocusOverlay />
	<WindowButtons />
</template>

<style scoped>
main {
	display: grid;
	grid-template:
		'.     logo      .' 3rem
		'.     .         .' 1fr
		'.     steps     .' 300px
		/ 1fr  600px     1fr;
	place-items: center;
	gap: var(--gap);
	height: 100%;
	min-height: 0;
}
</style>
