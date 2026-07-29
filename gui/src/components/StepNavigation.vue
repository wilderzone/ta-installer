<script setup lang="ts">
import { store } from '../store';
import { DirectoryState } from '../tai';
</script>

<template>
	<section class="step-navigation">
		<button
			title="Previous step"
			:disabled="store.step <= 1"
			@click="store.back()"
		>⋘</button>
		<button
			v-if="store.step < 3"
			title="Next step"
			:disabled="store.step >= 3"
			@click="store.next()"
		>⋙</button>
		<button
			v-else
			class="primary"
			:disabled="store.loading || store.pathState === DirectoryState.Invalid || store.isCustomRepo"
			@click="store.run()"
		>{{ store.action }}</button>
	</section>
</template>

<style scoped>
.step-navigation {
	display: flex;
	justify-content: space-between;
	align-self: flex-end;
	width: 100%;

	& button:not(.primary):disabled {
		opacity: 0;
	}
}
</style>
