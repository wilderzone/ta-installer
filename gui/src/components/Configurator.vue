<script setup lang="ts">
import { ref } from 'vue';
import { DEV } from '../dev';
import { type Action, TAI } from '../tai';
import LogLine from './LogLine.vue';
import InstallPathField from './InstallPathField.vue';
import ProgressBar from './ProgressBar.vue';
import RepositoryField from './RepositoryField.vue';

const defaultRepo = DEV ? 'http://localhost:3000/ta' : 'https://cdn.wilderzone.live/ta';

const action = ref<Action>('verify');
const path = ref('');
const repo = ref(defaultRepo);
const progressMax = ref(0);
const progressValue = ref(0);

function updateProgressBar(data: string | 0): void {
	if (data === 0) {
		progressMax.value = 0;
		progressValue.value = 0;
	} else {
		const result = /\[(\d+)\/(\d+)\]/g.exec(data);
		if (result && result[1] && result[2]) {
			progressMax.value = Number(result[2]);
			progressValue.value = Number(result[1]);
		}
	}
}

function onError(_data: string): void {
	updateProgressBar(0);
}

function onInfo(data: string): void {
	updateProgressBar(data);
}

async function run(): Promise<void> {
	await TAI.run(action.value, path.value, repo.value);
}

TAI.attach(onInfo, onError);
</script>

<template>
	<form class="column" action="" @submit.prevent>
		<InstallPathField v-model="path" />
		<RepositoryField v-model="repo" />
		<fieldset>
			<legend>Action</legend>
			<select v-model="action">
				<option value="install">Install</option>
				<option value="verify">Verify</option>
				<option value="update">Update</option>
				<option value="uninstall">Uninstall</option>
			</select>
		</fieldset>
		<fieldset>
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
		<div class="row">
			<ProgressBar :max="progressMax" :value="progressValue"/>
			<button type="submit" @click="run()">{{ action }}</button>
		</div>
	</form>
</template>

<style scoped>
fieldset:has(output) {
	height: 100%;
	min-height: 0;
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
	height: 1.5em;
	padding: 0;
	color: var(--color-text);
	font-size: 0.8rem;
	background-color: var(--color-background);
	border-radius: 50%;
}
</style>
