<script setup lang="ts">
import { ref } from 'vue';
import { DEV } from '../dev';
import { type Action, DirectoryState, TAI } from '../tai';
import LogLine from './LogLine.vue';
import InstallPathField from './InstallPathField.vue';
import ProgressBar from './ProgressBar.vue';
import RadioOption from './RadioOption.vue';
import RadioRow from './RadioRow.vue';
import RepositoryField from './RepositoryField.vue';

const defaultRepo = DEV ? 'http://localhost:3000/ta' : 'https://cdn.wilderzone.live/ta';

const action = ref<Action>('install');
const loading = ref(false);
const path = ref('');
const pathState = ref(DirectoryState.Invalid);
const refreshPath = ref(1);
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

function onComplete(_data: string): void {
	refreshPath.value++;
	loading.value = false;
}

function onPathUpdate(state: DirectoryState): void {
	pathState.value = state;
	switch (state) {
		case DirectoryState.Installed:
			action.value = 'verify';
			break;
		case DirectoryState.Valid:
			action.value = 'install';
			break;
	}
}

async function run(): Promise<void> {
	loading.value = true;
	await TAI.run(action.value, path.value, repo.value);
}

TAI.attach(onInfo, onError, onComplete);
</script>

<template>
	<form class="column" action="" @submit.prevent>
		<InstallPathField
			:disabled="loading"
			:refresh="refreshPath"
			@update="onPathUpdate($event)"
			v-model="path"
		/>
		<RepositoryField
			:disabled="loading"
			v-model="repo"
		/>
		<fieldset>
			<legend>Action</legend>
			<RadioRow
				:disabled="loading"
				name="action"
				v-model="action"
			>
				<RadioOption
					:disabled="pathState !== DirectoryState.Valid"
					label="Install"
					value="install"
				/>
				<RadioOption
					:disabled="pathState !== DirectoryState.Installed"
					label="Verify"
					value="verify"
				/>
				<RadioOption
					:disabled="pathState !== DirectoryState.Installed"
					label="Update"
					value="update"
				/>
				<RadioOption
					:disabled="pathState !== DirectoryState.Installed"
					label="Uninstall"
					value="uninstall"
				/>
			</RadioRow>
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
			<button
				type="submit"
				:disabled="loading || pathState === DirectoryState.Invalid"
				@click="run()"
			>{{ action }}</button>
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
