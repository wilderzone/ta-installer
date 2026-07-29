<script setup lang="ts">
import LogLine from './LogLine.vue';
import { TAI } from '../tai';
</script>

<template>
	<fieldset class="log">
		<legend>Log</legend>
		<button
			class="clear"
			title="Clear the log output"
			@click="TAI.clearLog()"
		>✕</button>
		<output>
			<LogLine v-for="line of TAI.log.value" v-bind="line" />
		</output>
	</fieldset>
</template>

<style scoped>
fieldset.log {
	height: 100%;
	min-height: 0;
	padding-inline: var(--fieldset-inline-padding);

	& legend {
		padding-inline: 0;
	}
}

output {
	display: flex;
	flex-flow: column-reverse nowrap;
	height: auto;
	max-height: 100%;
	margin-inline: calc(var(--fieldset-inline-padding) * -1);
	overflow: hidden auto;
}

button.clear {
	position: absolute;
	top: -1.1rem;
	right: var(--fieldset-inline-padding);
	display: grid;
	place-items: center;
	width: 1.5em;
	min-width: 1.5em;
	height: 1.5em;
	min-height: 1.5em;
	padding: 0;
	color: var(--color-text);
	font-size: 0.8rem;
	background-color: var(--color-card);
	border-radius: 50%;
}
</style>
