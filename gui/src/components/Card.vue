<script setup lang="ts">
import { computed } from 'vue';
import { store } from '../store';

defineProps<{
	name: string;
}>();

const glow = computed(() => store.progressIndeterminate || store.progressValue && !store.progressComplete);
</script>

<template>
	<section class="card" :name>
		<div class="background">
			<div class="shape blur"></div>
			<div class="shape colour"></div>
			<div class="line left" :class="{ glow }"></div>
			<div class="line right" :class="{ glow }"></div>
		</div>
		<slot></slot>
	</section>
</template>

<style scoped>
.card {
	--cutout: 12px;
	--inset: calc(var(--cutout) * -1);
	--offset: 12%;
	--line-color: #acf1;
	--line-color-glow: #acfd;
	width: 100%;
	height: 100%;
	padding: var(--gap-m);

	& > .background {
		position: absolute;
		inset: 0 var(--inset);
		display: grid;
		place-items: center;
	}
}

.shape {
	content: '';
	position: absolute;
	inset: 0;

	&::before {
		content: '';
		position: absolute;
		inset: 0;
		border-radius: calc(var(--cutout) + 13px);
		clip-path: polygon(
			calc(  0% + var(--cutout))        0%,
			calc(100% - var(--cutout))        0%,
			calc(100% - var(--cutout)) calc(  0% + var(--offset)),
					100%                  calc(  0% + var(--offset) + var(--cutout)),
					100%                  calc(100% - var(--offset) - var(--cutout)),
			calc(100% - var(--cutout)) calc(100% - var(--offset)),
			calc(100% - var(--cutout))      100%,
			calc(  0% + var(--cutout))      100%,
			calc(  0% + var(--cutout)) calc(100% - var(--offset)),
					0%                  calc(100% - var(--offset) - var(--cutout)),
					0%                  calc(  0% + var(--offset) + var(--cutout)),
			calc(  0% + var(--cutout)) calc(  0% + var(--offset))
		);
	}

	&.blur::before {
		backdrop-filter: blur(8px) brightness(0.3);
	}

	&.colour::before {
		background: var(--color-card);
		box-shadow:
			0 3px 4px -3px #ddd2 inset,
			0 -3px 4px -3px #000a inset;
	}

	&.colour {
		filter: drop-shadow(0 5px 12px #0006);
	}
}

.line {
	position: absolute;
	height: calc(100% - var(--offset) * 2 - var(--cutout) * 3);
	transition: 0.5s ease border-color, 0.2s ease box-shadow;

	&.left {
		left: 7px;
		border-left: 2px solid var(--line-color);
	}

	&.right {
		right: 7px;
		border-right: 2px solid var(--line-color);
	}

	&.glow {
		border-color: var(--line-color-glow);
		box-shadow:
			0 0 30px #fff,
			0 0 8px #fffb;
	}
}
</style>
