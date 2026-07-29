<script setup lang="ts">
import { store } from '../store';
</script>

<template>
	<section class="step-navigation">
		<button
			title="Previous step"
			:disabled="store.step <= 1"
			@click="store.back()"
		>⋘</button>
		<button
			title="Next step"
			:disabled="store.step >= 3"
			@click="store.next()"
		>⋙</button>
	</section>
</template>

<style scoped>
.step-navigation {
	display: flex;
	justify-content: space-between;
	align-self: flex-end;
	width: 100%;
}
</style>
