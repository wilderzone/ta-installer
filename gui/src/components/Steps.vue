<script setup lang="ts">
import { store } from '../store';
</script>

<template>
	<section class="steps">
		<ol>
			<li
				class="step"
				:class="{ current: store.step === 1, complete: store.step > 1 }"
				@click="store.step = 1"
			>
				<span class="number">1</span>
				<span class="check">✓</span>
				<span class="name">Game Type</span>
			</li>
			<li
				class="step"
				:class="{ current: store.step === 2, complete: store.step > 2 }"
				@click="store.step = 2"
			>
				<span class="number">2</span>
				<span class="check">✓</span>
				<span class="name">Location</span>
			</li>
			<li
				class="step"
				:class="{ current: store.step === 3, complete: store.step > 3 }"
				@click="store.step = 3"
			>
				<span class="number">3</span>
				<span class="check">✓</span>
				<span class="name">Install</span>
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
	min-width: 20ch;
	padding: calc(var(--gap) / 2);
	border-radius: var(--border-radius);
	user-select: none;

	&:where(:focus-visible, :hover) {
		background: var(--color-card);
	}

	&:where(.current) {
		background: var(--color-card);
	}

	& :is(.check, .number) {
		display: grid;
		place-items: center;
		flex-shrink: 0;
		width: 2em;
		height: 2em;
		border: 1px solid var(--color-border);
		border-radius: 50%;
	}

	& .check {
		display: none;
	}

	&.current .number {
		color: var(--color-text-on-primary);
		background: var(--color-primary);
		border-color: var(--color-primary);
	}

	&.complete .check {
		display: grid;
		color: var(--color-text-on-primary);
		background: var(--color-primary);
		border-color: var(--color-primary);
	}

	&.complete .number {
		display: none;
	}
}
</style>
