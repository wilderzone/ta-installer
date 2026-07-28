<script setup lang="ts">
import Background from './components/Background.vue';
import Card from './components/Card.vue';
import Configurator from './components/Configurator.vue';
import Logo from './components/Logo.vue';
import Scene from './components/Scene.vue';
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
</template>

<style scoped>
main {
	display: grid;
	grid-template:
		'.     logo      .' 3rem
		'.     steps     .' 1fr
		/ 1fr  515px      1fr;
	place-items: center;
	gap: var(--gap);
	height: 100%;
	min-height: 0;
}
</style>
