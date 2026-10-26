<script setup lang="ts">
import { ref } from 'vue';
import { loading, store } from '../store';
import { useFocusRecovery } from '../utilities';

const navigation = ref<HTMLElement>();
const previousButton = ref<HTMLButtonElement>();
const nextButton = ref<HTMLButtonElement>();
const cancelButton = ref<HTMLButtonElement>();
const runButton = ref<HTMLButtonElement>();

useFocusRecovery(navigation, [() => store.step, () => store.running, loading.value], () => (
	// Only Cancel is usable while loading, and the other buttons will change once it's done.
	loading.value.value
		? [cancelButton.value]
		: [cancelButton.value, runButton.value, nextButton.value, previousButton.value]
));
</script>

<template>
	<section ref="navigation" class="step-navigation">
		<button
			ref="previousButton"
			type="button"
			title="Previous step"
			aria-label="Previous step"
			:disabled="store.step <= 1"
			@click="store.back()"
		>⋘</button>
		<button
			v-if="store.step < 3"
			ref="nextButton"
			type="button"
			title="Next step"
			aria-label="Next step"
			:disabled="store.step >= 3"
			@click="store.next()"
		>⋙</button>
		<template v-else>
			<button
				v-if="store.running"
				ref="cancelButton"
				type="button"
				class="cancel"
				@click="store.cancel()"
			>Cancel</button>
			<button
				ref="runButton"
				type="submit"
				class="primary"
				:disabled="loading.value.value || store.step1Error || store.step2Error"
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
