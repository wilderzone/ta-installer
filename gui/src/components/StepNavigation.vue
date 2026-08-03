<script setup lang="ts">
import { loading, store } from '../store';
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
		<template v-else>
			<button
				v-if="store.running"
				class="cancel"
				@click="store.cancel()"
			>Cancel</button>
			<button
				class="primary"
				:disabled="loading.value.value || store.step1Error || store.step2Error"
				@click="store.run()"
			>{{ store.action }}</button>
		</template>
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

	& button.cancel {
		margin-left: auto;
	}
}
</style>
