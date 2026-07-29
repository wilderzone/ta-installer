<script setup lang="ts">
import LogLine from './LogLine.vue';
import { TAI } from '../tai';
</script>

<template>
	<section name="Log" class="log">
		<div class="border-top"></div>
		<div class="border-right"></div>
		<div class="border-bottom"></div>
		<div class="border-left"></div>
		<legend>Log</legend>
		<button
			class="clear"
			title="Clear the log output"
			@click="TAI.clearLog()"
		>✕</button>
		<output>
			<LogLine v-for="line of TAI.log.value" v-bind="line" />
		</output>
	</section>
</template>

<style scoped>
section.log {
	--padding: var(--fieldset-inline-padding);
	--padding-half: calc(var(--fieldset-inline-padding) / 2);
	display: grid;
	grid-template:
		'.     legend  top     clear   .    ' 0.5rem
		'left  legend  top     clear   right' 0.5rem
		'left  output  output  output  right' 1fr
		'left  output  output  output  right' var(--padding)
		'left  bottom  bottom  bottom  right' 0.5rem
		/ var(--padding) auto 1fr auto var(--padding);
	align-items: center;
	width: 100%;
	height: 100%;
	min-height: 56px;
	font-size: 0.9rem;

	& legend {
		grid-area: legend;
		padding-inline: var(--padding-half);
	}
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
	grid-area: clear;
	display: grid;
	place-items: center;
	width: 1.5em;
	min-width: 1.5em;
	height: 1.5em;
	min-height: 1.5em;
	padding: 0;
	color: var(--color-text);
	font-size: 0.8rem;
	background: unset;
	border-radius: 50%;
}

.border-top {
	grid-area: top;
	width: 100%;
	height: 1px;
	background: var(--color-border);
}
.border-right {
	grid-area: right;
	width: 100%;
	height: 100%;
	border: 1px solid var(--color-border);
	border-left: unset;
	border-radius: 0 var(--border-radius) var(--border-radius) 0;
}
.border-bottom {
	grid-area: bottom;
	align-self: flex-end;
	width: 100%;
	height: 1px;
	background: var(--color-border);
}
.border-left {
	grid-area: left;
	width: 100%;
	height: 100%;
	border: 1px solid var(--color-border);
	border-right: unset;
	border-radius: var(--border-radius) 0 0 var(--border-radius);
}
</style>
