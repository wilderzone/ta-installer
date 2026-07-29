<script setup lang="ts">
import { store } from '../store';
</script>

<template>
	<section class="steps">
		<ol>
			<li
				class="step"
				:class="{ current: store.step === 1, complete: store.step > 1, error: store.step1Error }"
				@click="store.step = 1"
			>
				<span class="number">1</span>
				<span class="check">✓</span>
				<span class="cross">✕</span>
				<span class="name">Game Type</span>
			</li>
			<li
				class="step"
				:class="{ current: store.step === 2, complete: store.step > 2, error: store.step2Error }"
				@click="store.step = 2"
			>
				<span class="number">2</span>
				<span class="check">✓</span>
				<span class="cross">✕</span>
				<span class="name">Location</span>
			</li>
			<li
				class="step"
				:class="{ current: store.step === 3, complete: store.step > 3, error: store.step3Error }"
				@click="store.step = 3"
			>
				<span class="number">3</span>
				<span class="check">✓</span>
				<span class="cross">✕</span>
				<span class="name">{{ store.action }}</span>
			</li>
		</ol>
	</section>
</template>

<style scoped>
.steps {
	width: 100%;

	& ol {
		display: flex;
		justify-content: space-between;
		gap: var(--gap);
		margin: 0;
		padding: 0;
		list-style: none;
	}
}

.step {
	display: flex;
	align-items: center;
	gap: var(--gap);
	width: 100%;
	min-width: 20ch;
	padding: calc(var(--gap) / 2);
	text-transform: capitalize;
	border-radius: var(--border-radius);

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
	}
}
</style>
