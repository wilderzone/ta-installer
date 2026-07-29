<script setup lang="ts">
import { open } from '@tauri-apps/plugin-dialog';
import { computed, nextTick, ref, watch } from 'vue';
import { DirectoryState, TAI } from '../tai';

const props = defineProps<{ disabled: boolean, refresh: number }>();
const model = defineModel<string>({ required: true });
const emit = defineEmits<{ update: [state: DirectoryState] }>();

const installationFound = ref(false);
const noPath = computed(() => !model.value);

async function browse(): Promise<void> {
	model.value = await open({
		multiple: false,
		directory: true,
	}) || model.value;

	await check();
}

async function check(): Promise<void> {
	nextTick(async () => {
		if (!model.value) {
			emit('update', DirectoryState.Invalid);
			return;
		}
		installationFound.value = await TAI.checkPath(model.value);
		emit('update', installationFound.value ? DirectoryState.Installed : DirectoryState.Valid);
	});
}

watch(() => props.refresh, () => void check());
</script>

<template>
	<fieldset class="borderless">
		<legend>Select a folder in which to install the game files.</legend>
		<div class="row">
			<input type="text" :disabled @change="check()" v-model="model">
			<button type="button" :disabled @click="browse()">Browse</button>
		</div>
		<p v-if="noPath" class="red">No path specified.</p>
		<p v-else-if="installationFound" class="green">Installation found.</p>
		<p v-else class="green">Path valid.</p>
	</fieldset>
</template>

<style scoped>
</style>
