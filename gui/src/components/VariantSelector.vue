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

async function check(): Promise<void> {
	nextTick(async () => {
		if (model.value === 'custom') {
			isCustom.value = true;
			await openUrl(repository.custom);
			return;
		}

		loading.value = true;
		const repo = model.value;
		const found = await TAI.checkRepo(repo);
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
	<fieldset>
		<legend>Variant</legend>
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
		<p v-else-if="repositoryFound" class="green">Repository found.</p>
		<p v-else-if="!repositoryFound" class="red">Repository not found. Please check your internet connection.</p>
	</fieldset>
</template>

<style scoped>
</style>
