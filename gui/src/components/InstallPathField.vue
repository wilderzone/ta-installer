<script setup lang="ts">
import { open } from '@tauri-apps/plugin-dialog';
import { computed, nextTick, ref } from 'vue';
import { TAI } from '../tai';

const props = defineProps<{ disabled: boolean }>();
const model = defineModel<string>({ required: true });

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
		installationFound.value = await TAI.checkPath(model.value);
	});
}
</script>

<template>
	<fieldset>
		<legend>Install Path</legend>
		<div class="row">
			<input type="text" :disabled @change="check()" v-model="model">
			<button :disabled @click="browse()">Browse</button>
		</div>
		<p v-if="noPath" class="red">No path specified.</p>
		<p v-else-if="installationFound" class="green">Installation found.</p>
		<p v-else class="green">Path valid.</p>
	</fieldset>
</template>

<style scoped>
p {
	margin: 0;
	padding: var(--gap) var(--border-radius) 0;
	font-size: 0.8rem;
	font-style: italic;

	&.green {
		color: var(--color-text-green);
	}

	&.red {
		color: var(--color-text-red);
	}
}
</style>
