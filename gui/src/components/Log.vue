<script setup lang="ts">
import LogLine from './LogLine.vue';
import { LogLineData, TAI } from '../tai';

const placeholder: LogLineData = {
	type: 'info',
	message: '\[2mProgress information will appear here...\[0m'
};
</script>

<template>
	<section name="Log" class="log">
		<output aria-label="Progress log">
			<LogLine v-if="!TAI.log.value.length" v-bind="placeholder" />
			<LogLine v-for="line of TAI.log.value" v-bind="line" />
		</output>
		<button
			class="clear"
			title="Clear the progress log"
			aria-label="Clear the progress log"
			@click="TAI.clearLog()"
		>✕</button>
	</section>
</template>

<style scoped>
section.log {
	--padding: var(--fieldset-inline-padding);
	--padding-half: calc(var(--fieldset-inline-padding) / 2);
	display: flex;
	flex-direction: column;
	height: 100%;
	min-height: 56px;
	margin: 0 var(--border-radius);
	padding: var(--padding);
	padding-top: var(--padding-half);
	font-size: 0.9rem;
	background: var(--color-card);
	border-radius: 0 0 var(--border-radius) var(--border-radius);
}

output {
	grid-area: output;
	display: flex;
	flex-flow: column-reverse nowrap;
	height: auto;
	max-height: 100%;
	margin-inline: calc(var(--padding) * -1);
	overflow: hidden auto;
}

button.clear {
	position: absolute;
	top: var(--padding-half);
	right: var(--padding);
	display: grid;
	place-items: center;
	width: 2em;
	min-width: 2em;
	height: 2em;
	min-height: 2em;
	padding: 0;
	color: var(--color-text);
	font-size: 0.8rem;
	background: var(--color-card);
	border-radius: 50%;
}
</style>
