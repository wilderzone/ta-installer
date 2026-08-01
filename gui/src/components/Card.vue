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
			<div class="metal left" :class="{ glow }"></div>
			<div class="metal right" :class="{ glow }"></div>
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
			0 0 50px #fff,
			0 0 30px #fff,
			0 0 8px #fffb,
			0 0 2px #fff8;
	}
}

@property --metal-lightness {
	syntax: '<integer>';
	inherits: false;
	initial-value: 20;
}

.metal {
	--line-color: #222;
	--metal-color: #484848;
	--metal-lightness: 20;
	--metal-lightness-glow: 40;

	position: absolute;
	top: 3px;
	bottom: 3px;
	display: grid;
	place-items: center;
	width: calc(var(--cutout) * 1.5);
	background: radial-gradient(circle at 100% 50%, hsl(from var(--metal-color) h s calc(l + var(--metal-lightness))), var(--metal-color) 125%);
	border-top-left-radius: 50% 38px;
	border-top-right-radius: 120% 6px;
	border-bottom-right-radius: 120% 6px;
	border-bottom-left-radius: 50% 38px;
	transition: 0.5s ease --metal-lightness;
	clip-path: polygon(
			   0%                  calc(  0% + var(--cutout)),
			 100%                         0%,
			 100%                  calc(  0% + var(--offset)),
		calc(100% - var(--cutout)) calc(  0% + var(--offset) + var(--cutout)),
		calc(100% - var(--cutout)) calc(100% - var(--offset) - var(--cutout)),
			 100%                  calc(100% - var(--offset)),
			 100%                       100%,
			   0%                  calc(100% - var(--cutout))
	);

	&.left {
		left: calc(var(--cutout) * -0.5);
	}

	&.right {
		right: calc(var(--cutout) * -0.5);
		transform: rotateZ(180deg);
	}

	&.glow {
		--metal-lightness: var(--metal-lightness-glow);
	}

	&::before,
	&::after {
		--dash-offset: 2.5%;
		content: '';
		position: absolute;
		width: 3px;
		height: calc(var(--offset) - var(--dash-offset) * 2);
		background: var(--line-color);
		border-radius: 10px;
	}

	&::before {
		top: calc(var(--dash-offset) + var(--cutout) / 2);
	}

	&::after {
		bottom: calc(var(--dash-offset) + var(--cutout) / 2);
	}
}
</style>
