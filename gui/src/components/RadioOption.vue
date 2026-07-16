<script setup lang="ts">
import { computed, inject } from 'vue';

const props = defineProps<{ disabled?: boolean | string; label: string; value: string; }>();
const selected = inject<string>('radio-selected');
const select = inject<(value: string) => void>('radio-select');
const disabled = inject<boolean>('radio-disabled');
const hash = inject<string>('radio-hash');
const name = inject<string>('radio-name');
const disabledMessage = computed(() => typeof props.disabled === 'string' ? props.disabled : undefined);

function change(event: Event, value: string): void {
	const target = event.target;
	if (!(target instanceof HTMLInputElement)) return;
	if (!target.checked) return;
	if (props.disabled) return;
	select?.(value);
}
</script>

<template>
	<label
		class="radio-option"
		:for="`${hash}-${value}`"
		:title="disabledMessage"
	>
		<span>{{ label }}</span>
		<input
			type="radio"
			:id="`${hash}-${value}`"
			:name
			:disabled="disabled || Boolean(props.disabled)"
			:checked="selected === value"
			@change="change($event, value)"
		>
	</label>
</template>

<style scoped>
.radio-option {
	position: relative;
	display: grid;
	place-items: center;
	width: 100%;
	height: 2em;
	background-color: #8884;

	&:first-of-type {
		border-radius: var(--border-radius) 0 0 var(--border-radius);
	}

	&:last-of-type {
		border-radius: 0 var(--border-radius) var(--border-radius) 0;
	}

	& > * {
		position: absolute;
	}

	&:has(input:checked) {
		color: var(--color-text-on-primary);

		&::before {
			--inset: 2px;
			content: '';
			position: absolute;
			inset: var(--inset);
			background-color: var(--color-primary);
			border-radius: calc(var(--border-radius) - var(--inset));
		}
	}

	& span {
		z-index: 1;
		user-select: none;
	}

	& input {
		inset: 0;
		margin: 0;
		opacity: 0;
	}

	&:has(input:disabled) {
		filter: grayscale(1);
		/* pointer-events: none; */

		&::before,
		& span {
			filter: brightness(0.6);
		}
	}
}
</style>
