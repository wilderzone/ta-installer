<script setup lang="ts">
import { openUrl } from '@tauri-apps/plugin-opener';
import { nextTick, ref, watch } from 'vue';
import { TAI } from '../tai';
import RadioRow from './RadioRow.vue';
import RadioOption from './RadioOption.vue';
import { repository } from '../config.ts';

defineProps<{ disabled: boolean }>();
const model = defineModel<string>({ required: true });

const loading = ref(false);
const isCustom = ref(false);
const repositoryFound = ref(false);

async function delay(duration: number): Promise<void> {
	return new Promise(resolve => {
		window.setTimeout(resolve, duration);
	});
}

async function check(): Promise<void> {
	nextTick(async () => {
		if (model.value === 'custom') {
			isCustom.value = true;
			await openUrl(repository.custom);
			return;
		}

		loading.value = true;
		const repo = model.value;
		const found = await TAI.checkRepo(repository[repo]);
		await delay(500);
		if (repo === model.value) {
			repositoryFound.value = found;
			loading.value = false;
		}
	});
}

watch(model, () => {
	isCustom.value = false;
	repositoryFound.value = false;
	void check();
}, { immediate: true });
</script>

<template>
	<fieldset class="borderless">
		<legend>Game Type</legend>
		<RadioRow
			:disabled
			name="variant"
			v-model="model"
		>
			<RadioOption
				label="Vanilla"
				value="vanilla"
			/>
			<RadioOption
				label="Modded"
				value="modded"
			/>
			<RadioOption
				label="Custom"
				value="custom"
			/>
		</RadioRow>
		<p v-if="loading" class="grey">Checking...</p>
		<p v-else-if="isCustom" class="grey">To install a custom variant, please follow the <a :href="repository.custom" target="_blank" rel="noreferrer">TAI CLI instructions</a>.</p>
		<p v-else-if="repositoryFound" class="green">Game files available.</p>
		<p v-else-if="!repositoryFound" class="red">Game files could not be found. Please check your internet connection.</p>
	</fieldset>
</template>

<style scoped>
</style>
