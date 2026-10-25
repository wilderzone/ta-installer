<script setup lang="ts">
import { computed, inject, Ref } from 'vue';

const props = defineProps<{
	disabled?: boolean | string;
	image?: string;
	label?: string;
	sublabel?: string;
	value: string;
}>();
const selected = inject<string>('radio-selected');
const select = inject<(value: string) => void>('radio-select');
const disabled = inject<Ref<boolean>>('radio-disabled');
const hash = inject<string>('radio-hash');
const name = inject<Ref<string>>('radio-name');
const disabledMessage = computed(() => {
	const rowDisabled = disabled?.value;
	if (typeof props.disabled === 'string' && !rowDisabled) {
		return props.disabled;
	} else {
		return undefined
	}
});

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
		<img v-if="image" :src="image" height="48" alt="" draggable="false">
		<div class="text">
			<span class="label"><slot name="label">{{ label }}</slot></span>
			<span class="sublabel"><slot name="sublabel">{{ sublabel }}</slot></span>
		</div>
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
	display: flex;
	flex-direction: column;
	justify-content: center;
	align-items: center;
	gap: calc(var(--gap-m) / 5);
	width: 100%;
	height: 7em;
	background-color: var(--color-input);

	&:first-of-type {
		border-radius: var(--border-radius) 0 0 var(--border-radius);
	}

	&:last-of-type {
		border-radius: 0 var(--border-radius) var(--border-radius) 0;
	}

	&:where(:focus-visible, :hover):not(:has(input:disabled)) {
		&::before {
			--inset: 2px;
			content: '';
			position: absolute;
			inset: var(--inset);
			background: var(--color-input);
			border-radius: calc(var(--border-radius) - var(--inset));
		}
	}

	&:has(input:checked) {
		color: var(--color-text-on-primary);

		&::before {
			--inset: 2px;
			content: '';
			position: absolute;
			inset: var(--inset);
			background: var(--color-primary);
			border-radius: calc(var(--border-radius) - var(--inset));
		}

		& .sublabel {
			color: var(--color-text-dim-on-primary);
		}
	}

	& img {
		filter: saturate(0.85) brightness(1.2);
	}

	& .text {
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;

		& .sublabel {
			color: var(--color-text-dim);
			font-size: 0.8rem;
		}
	}

	& input {
		position: absolute;
		inset: 0;
		margin: 0;
		opacity: 0;
		pointer-events: none;
	}

	&:has(input:disabled) {
		filter: grayscale(1);

		&::before,
		& span {
			filter: brightness(0.6);
		}
	}
}
</style>
