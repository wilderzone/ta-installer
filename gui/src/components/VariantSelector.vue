<script setup lang="ts">
import { ask } from '@tauri-apps/plugin-dialog';
import { openUrl } from '@tauri-apps/plugin-opener';
import { nextTick, ref, watch } from 'vue';
import { TAI } from '../tai';
import RadioRow from './RadioRow.vue';
import RadioOption from './RadioOption.vue';
import { customInstructions, repository } from '../config.ts';
import { store } from '../store.ts';

defineProps<{ disabled: boolean }>();
const model = defineModel<string>({ required: true });
const isCustom = ref(false);
const repositoryFound = ref(false);

async function delay(duration: number): Promise<void> {
	return new Promise(resolve => {
		window.setTimeout(resolve, duration);
	});
}

async function openCustomInstructions(): Promise<void> {
	const message = 'For custom Tribes: Ascend installations, please use the command-line interface. '
	              + '\nWould you like to open the CLI instructions?';
	const proceed = await ask(message, { title: 'Custom Tribes: Ascend Installation', kind: 'info' });
	if (proceed) {
		await openUrl(customInstructions);
	}
}

async function check(): Promise<void> {
	nextTick(async () => {
		if (model.value === 'custom') {
			isCustom.value = true;
			store.step1Error = true;
			await openCustomInstructions();
			return;
		}

		store.loading = true;
		const repo = model.value;
		const found = await TAI.checkRepo(repository);
		await delay(500);
		if (repo === model.value) {
			store.step1Error = !found;
			repositoryFound.value = found;
			store.loading = false;
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
		<RadioRow
			:disabled
			name="variant"
			v-model="model"
		>
			<RadioOption
				image="124.png"
				label="Vanilla"
				sublabel="(4.5GB)"
				value="vanilla"
			/>
			<RadioOption
				image="129.png"
				label="Modded"
				sublabel="(8GB)"
				value="modded"
			/>
			<RadioOption
				image="118.png"
				label="Custom"
				value="custom"
			/>
		</RadioRow>
		<p v-if="store.loading" class="grey">Checking...</p>
		<p v-else-if="isCustom" class="grey">To install a custom variant of Tribes: Ascend, please follow the <a :href="customInstructions" target="_blank" rel="noreferrer">TAI CLI instructions</a>.</p>
		<p v-else-if="repositoryFound" class="green">Game files available.</p>
		<p v-else-if="!repositoryFound" class="red">Game files could not be found. Please check your internet connection.</p>
	</fieldset>
</template>

<style scoped>
</style>
