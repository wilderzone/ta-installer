<script setup lang="ts">
import AssetList from '../components/AssetList.vue';
import AssetNavigation from '../components/AssetNavigation.vue';
import Background from '../components/Background.vue';
import FocusOverlay from '../components/FocusOverlay.vue';
</script>

<template>
	<Background />
	<main>
		<AssetList v-grid:list />
		<AssetNavigation />
	</main>
	<FocusOverlay />
</template>

<style scoped>
main {
	position: fixed;
	inset: 0;
	display: flex;
	flex-direction: column;
	height: 100%;
	min-height: 0;

	& img {
		grid-area: icon;
		height: 2.4rem;
	}

	& h1 {
		grid-area: heading;
		align-self: center;
		margin: 0;
		font-size: 1.4rem;
		font-style: italic;
		font-weight: normal;
	}
}
</style>
