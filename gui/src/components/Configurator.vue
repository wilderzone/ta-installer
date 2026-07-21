<script setup lang="ts">
import { computed, ref } from 'vue';
import { repository } from '../config.ts';
import { type Action, DirectoryState, TAI } from '../tai';
import LogLine from './LogLine.vue';
import InstallPathField from './InstallPathField.vue';
import ProgressBar from './ProgressBar.vue';
import RadioOption from './RadioOption.vue';
import RadioRow from './RadioRow.vue';
import VariantSelector from './VariantSelector.vue';

const action = ref<Action>('install');
const loading = ref(false);
const path = ref('');
const pathState = ref(DirectoryState.Invalid);
const refreshPath = ref(1);
const repo = ref('vanilla');
const progressMax = ref(0);
const progressValue = ref(0);
const isCustomRepo = computed(() => repo.value === 'custom');

const installButtonDisabled = computed(() => {
	switch (pathState.value) {
		case DirectoryState.Valid:
			return false;
		case DirectoryState.Invalid:
			return 'Select a valid installation path before installing.';
		case DirectoryState.Installed:
			return 'An installation is already present at the selected path.';
	}
});

const verifyButtonDisabled = computed(() => {
	switch (pathState.value) {
		case DirectoryState.Installed:
			return false;
		default:
			return 'You need to install the game before it can be verified.';
	}
});

const updateButtonDisabled = computed(() => {
	switch (pathState.value) {
		case DirectoryState.Installed:
			return false;
		default:
			return 'You need to install the game before it can be updated.';
	}
});

const uninstallButtonDisabled = computed(() => {
	switch (pathState.value) {
		case DirectoryState.Installed:
			return false;
		default:
			return 'You need to install the game before it can be uninstalled.';
	}
});

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
			if (action.value === 'install') {
				action.value = 'verify';
			}
			break;
		case DirectoryState.Valid:
			action.value = 'install';
			break;
	}
}

async function run(): Promise<void> {
	loading.value = true;
	await TAI.run(action.value, path.value, repository[repo.value]);
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
		<VariantSelector
			:disabled="loading"
			v-model="repo"
		/>
		<fieldset class="borderless">
			<legend>Action</legend>
			<RadioRow
				:disabled="loading || isCustomRepo"
				name="action"
				v-model="action"
			>
				<RadioOption
					:disabled="installButtonDisabled"
					label="Install"
					value="install"
				/>
				<RadioOption
					:disabled="verifyButtonDisabled"
					label="Verify"
					value="verify"
				/>
				<RadioOption
					:disabled="updateButtonDisabled"
					label="Update"
					value="update"
				/>
				<RadioOption
					:disabled="uninstallButtonDisabled"
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
				:disabled="loading || pathState === DirectoryState.Invalid || isCustomRepo"
				@click="run()"
			>{{ action }}</button>
		</div>
	</form>
</template>

<style scoped>
fieldset:has(output) {
	height: 100%;
	min-height: 0;
	margin-top: var(--gap);
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
	height: 1.5em;
	min-height: 1.5em;
	padding: 0;
	color: var(--color-text);
	font-size: 0.8rem;
	background-color: var(--color-background);
	border-radius: 50%;
}
</style>
