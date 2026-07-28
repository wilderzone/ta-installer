<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
	value: number;
	max: number
	disabled?: boolean;
	indeterminate?: boolean;
	label?: string;
}>();

const valuePercentage = computed(() => props.value / props.max * 100);
</script>

<template>
	<div class="progress-bar" :class="{ complete: value >= max, disabled, indeterminate }">
		<progress :max :value="value">{{ value.toFixed(0) }}%</progress>
		<div class="track">
			<div class="bar" :style="{ width: indeterminate ? '100%' : `${valuePercentage}%` }"></div>
			<div v-if="label" class="percentage">{{ label }}</div>
			<div v-else class="percentage">{{ (valuePercentage || 0).toFixed(0) }}%</div>
		</div>
	</div>
</template>

<style scoped>
.progress-bar {
	width: 100%;
	height: 100%;

	& progress {
		position: absolute;
		left: 0;
		opacity: 0;
	}

	& .track {
		display: grid;
		align-items: center;
		width: 100%;
		height: 100%;
		border: 1px solid var(--color-border);
		border-radius: var(--border-radius);
		overflow: hidden;
	}

	& .bar {
		position: absolute;
		width: 0;
		height: 100%;
		background:
			0 0 / 30px 50% linear-gradient(65deg, var(--color-secondary) 10px, var(--color-tertiary) 10px, var(--color-tertiary) 24px, var(--color-secondary) 24px) repeat-x,
			0 100% / 30px 50% linear-gradient(115deg, var(--color-secondary) 10px, var(--color-tertiary) 10px, var(--color-tertiary) 24px, var(--color-secondary) 24px) repeat-x;
		filter: grayscale(0);
		animation: progress 0.4s linear infinite;
		transition: 0.1s ease filter;
	}

	& .percentage {
		position: absolute;
		max-width: 100%;
		padding: 0 1ch;
		color: white;
		font-weight: 300;
		font-style: italic;
		text-overflow: ellipsis;
		white-space: nowrap;
		overflow: hidden;
	}
}

.progress-bar.complete:not(.indeterminate) {
	& .bar {
		background: var(--color-secondary);
	}
}

.progress-bar.disabled {
	& .percentage {
		display: none;
	}

	& .bar {
		filter: grayscale(1);
	}
}

@keyframes progress {
	0%   { background-position: 0px 0, 0px 100%; }
	100% { background-position: 30px 0, 30px 100%; }
}
</style>
