<script setup lang="ts">
import { computed, provide } from 'vue';

const props = defineProps<{ disabled: boolean, name: string }>();
const model = defineModel<string>({ required: true });
const hash = crypto.randomUUID();
const disabled = computed(() => props.disabled);
const name = computed(() => props.name);

provide('radio-select', select);
provide('radio-selected', model);
provide('radio-disabled', disabled);
provide('radio-hash', hash);
provide('radio-name', name);

function select(value: unknown): void {
	if (typeof value === 'string') {
		model.value = value;
	}
}
</script>

<template>
	<div role="radiogroup" class="radio-row">
		<slot></slot>
	</div>
</template>

<style scoped>
.radio-row {
	display: flex;
	border-radius: var(--border-radius);

	&:has(:focus-visible) {
		outline: 1px solid var(--color-text);
	}
}
</style>
