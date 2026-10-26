<script setup lang="ts">
import { computed, ref, watch } from 'vue';

const props = defineProps<{
	ariaLabel?: string;
	value: number;
	max: number
	detail?: string;
	disabled?: boolean;
	indeterminate?: boolean;
	label?: string;
}>();

const ease = ref(false);
const complete = computed(() => props.value >= props.max && props.value);
const valuePercentage = computed(() => props.max ? props.value / props.max * 100 : 0);
const valueText = computed(() => {
	if (props.indeterminate) return undefined;
	if (props.label) return props.label;
	const percentage = `${(valuePercentage.value || 0).toFixed(0)}%`;
	return props.detail ? `${percentage}, ${props.detail}` : percentage;
});

// The bar should ease when moving forwards, but snap when moving backwards (ie, resetting).
watch(valuePercentage, (value, previous) => {
	ease.value = value > previous;
});
</script>

<template>
	<div class="progress-bar" :class="{ complete, disabled, ease, indeterminate }">
		<progress
			:aria-label
			:aria-valuetext="valueText"
			:max
			:value="indeterminate ? undefined : value"
		>{{ value.toFixed(0) }}%</progress>
		<div class="track" aria-hidden="true">
			<div class="bar" :style="{ width: indeterminate ? '100%' : `${valuePercentage}%` }"></div>
			<div class="bar glow" :style="{ width: indeterminate ? '100%' : `${valuePercentage}%` }"></div>
			<div v-if="label" class="percentage">{{ label }}</div>
			<div v-if="!label && !indeterminate" class="percentage">
				<span>{{ (valuePercentage || 0).toFixed(0) }}%</span>
				<span v-if="detail" class="detail">{{ detail }}</span>
			</div>
		</div>
	</div>
</template>

<style scoped>
.progress-bar {
	--bar-transition: 0s linear width;
	width: 100%;
	height: 100%;
	min-height: var(--input-height);

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
		transition: 0.1s ease filter, var(--bar-transition);

		@media (prefers-reduced-motion: reduce) {
			animation: none;
		}

		&.glow {
			filter: blur(5px) brightness(1.5) opacity(0.35);
			mix-blend-mode: plus-lighter;
			transition: 0.1s ease opacity, var(--bar-transition);
		}
	}

	& .percentage {
		--inset: 5px;
		position: absolute;
		top: var(--inset);
		bottom: var(--inset);
		left: var(--inset);
		display: flex;
		align-items: center;
		max-width: calc(100% - var(--inset) * 2);
		padding: 0 var(--inset);
		color: white;
		font-weight: 300;
		font-style: italic;
		text-overflow: ellipsis;
		white-space: nowrap;
		background: var(--color-card);
		border-radius: 2px;
		overflow: hidden;

		& .detail {
			min-width: 0;
			opacity: 0.7;

			&::before {
				content: '•';
				margin: 0 1ch;
			}
		}
	}
}

.progress-bar.ease:not(.indeterminate) {
	--bar-transition: 0.2s linear width;
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

		&.glow {
			opacity: 0;
		}
	}
}

@keyframes progress {
	0%   { background-position: 0px 0, 0px 100%; }
	100% { background-position: 30px 0, 30px 100%; }
}
</style>
