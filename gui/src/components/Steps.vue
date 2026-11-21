<script setup lang="ts">
import { store } from '../store';

function stepLabel(step: number, name: string, complete: boolean, error: boolean): string {
	let label = `Step ${step}: ${name}`;
	if (error) label += ', error';
	else if (complete) label += ', complete';
	return label;
}

/** Select a step and move focus to its tab, wrapping around at either end. */
function focusStep(step: number): void {
	if (step < 1) step = 3;
	if (step > 3) step = 1;
	store.step = step;
	document.getElementById(`tab-step-${step}`)?.focus();
}
</script>

<template>
	<section class="steps">
		<ol role="tablist" aria-label="Installation configuration steps">
			<li
				role="tab"
				id="tab-step-1"
				aria-controls="panel-step-1"
				:aria-selected="store.step === 1"
				:aria-label="stepLabel(1, 'Game Type', store.step > 1, store.step1Error)"
				:tabindex="store.step === 1 ? 0 : -1"
				class="step"
				:class="{ current: store.step === 1, complete: store.step > 1, error: store.step1Error }"
				@click="store.step = 1"
				@keydown.enter="store.step = 1"
				@keydown.space="store.step = 1"
				@keydown.left.prevent="focusStep(store.step - 1)"
				@keydown.right.prevent="focusStep(store.step + 1)"
				@keydown.home.prevent="focusStep(1)"
				@keydown.end.prevent="focusStep(3)"
			>
				<span class="number" aria-hidden="true">1</span>
				<span class="check" aria-hidden="true">✓</span>
				<span class="cross" aria-hidden="true">!</span>
				<span class="name" aria-hidden="true">Game Type</span>
			</li>
			<li
				role="tab"
				id="tab-step-2"
				aria-controls="panel-step-2"
				:aria-selected="store.step === 2"
				:aria-label="stepLabel(2, 'Location', store.step > 2, store.step2Error)"
				:tabindex="store.step === 2 ? 0 : -1"
				class="step"
				:class="{ current: store.step === 2, complete: store.step > 2, error: store.step2Error }"
				@click="store.step = 2"
				@keydown.enter="store.step = 2"
				@keydown.space="store.step = 2"
				@keydown.left.prevent="focusStep(store.step - 1)"
				@keydown.right.prevent="focusStep(store.step + 1)"
				@keydown.home.prevent="focusStep(1)"
				@keydown.end.prevent="focusStep(3)"
			>
				<span class="number" aria-hidden="true">2</span>
				<span class="check" aria-hidden="true">✓</span>
				<span class="cross" aria-hidden="true">!</span>
				<span class="name" aria-hidden="true">Location</span>
			</li>
			<li
				role="tab"
				id="tab-step-3"
				aria-controls="panel-step-3"
				:aria-selected="store.step === 3"
				:aria-label="stepLabel(3, store.action, store.step > 3, store.step3Error)"
				:tabindex="store.step === 3 ? 0 : -1"
				class="step"
				:class="{ current: store.step === 3, complete: store.step > 3, error: store.step3Error }"
				@click="store.step = 3"
				@keydown.enter="store.step = 3"
				@keydown.space="store.step = 3"
				@keydown.left.prevent="focusStep(store.step - 1)"
				@keydown.right.prevent="focusStep(store.step + 1)"
				@keydown.home.prevent="focusStep(1)"
				@keydown.end.prevent="focusStep(3)"
			>
				<span class="number" aria-hidden="true">3</span>
				<span class="check" aria-hidden="true">✓</span>
				<span class="cross" aria-hidden="true">!</span>
				<span class="name" aria-hidden="true">{{ store.action }}</span>
			</li>
		</ol>
	</section>
</template>

<style scoped>
.steps {
	width: 100%;
	min-width: 0;

	& ol {
		display: flex;
		justify-content: space-between;
		gap: var(--gap-m);
		margin: 0;
		padding: 0;
		list-style: none;
	}
}

.step {
	display: flex;
	align-items: center;
	gap: var(--gap-m);
	width: 100%;
	min-width: 0;
	padding: var(--gap-s);
	text-transform: capitalize;
	border-radius: var(--border-radius);

	&:focus-visible {
		outline: var(--focus-outline);
	}

	&:where(:focus-visible, :hover) {
		background: var(--color-card);
	}

	&:where(.current) {
		background: var(--color-card);
	}

	& :is(.check, .cross, .number) {
		display: grid;
		place-items: center;
		flex-shrink: 0;
		width: 2em;
		height: 2em;
		border: 1px solid var(--color-border);
		border-radius: 50%;
	}

	& :is(.check, .cross) {
		display: none;
		padding-bottom: 0.14em;
	}

	&.current .number {
		color: var(--color-text-on-primary);
		background: var(--color-primary);
		border-color: var(--color-primary);
	}

	&.complete .check {
		display: grid;
		color: var(--color-text-on-primary);
		background: var(--color-success);
		border-color: var(--color-success);
	}

	&.complete .number {
		display: none;
	}

	&.error :is(.check, .number) {
		display: none;
	}

	&.error .cross {
		display: grid;
		color: var(--color-text-on-primary);
		background: var(--color-error);
		border-color: var(--color-error);
		font-weight: bold;
	}
}
</style>
