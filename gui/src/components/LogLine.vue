<script setup lang="ts">
import { computed } from 'vue';
import { LogLineData } from '../tai';

const line = defineProps<LogLineData>();

const message = computed(() => {
	const html = line.message
		.replace(/\[2m([^]+)\[0m/g, '<span class="grey">$1</span>')
		.replace(/\[31m([^]+)\[0m/g, '<span class="red">$1</span>')
		.replace(/\[32m([^]+)\[0m/g, '<span class="green">$1</span>');

	return html;
});
</script>

<template>
	<p class="logline" :class="line.type" v-html="message"></p>
</template>

<style scoped>
.logline {
	flex-shrink: 0;
	margin: 0;
	min-height: calc(1em + 2px);
	padding: 2px var(--fieldset-inline-padding);
	font-size: 0.8rem;

	&.error {
		background-color: #f008;
	}
}
</style>

<style>
.logline span {
	&.grey {
		opacity: 0.5;
	}

	&.green {
		color: var(--color-text-green);
	}

	&.red {
		color: var(--color-text-red);
	}
}
</style>
