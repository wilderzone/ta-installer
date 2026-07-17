<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue';
import { TAI } from '../tai';

defineProps<{ disabled: boolean }>();
const model = defineModel<string>({ required: true });

const loading = ref(false);
const repositoryFound = ref(false);
const noUrl = computed(() => !model.value);

async function check(): Promise<void> {
	nextTick(async () => {
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
	repositoryFound.value = false;
});

onMounted(() => {
	void check();
});
</script>

<template>
	<fieldset>
		<legend>Repository</legend>
		<input type="text" :disabled @change="check()" v-model="model">
		<p v-if="loading" class="grey">Checking...</p>
		<p v-else-if="noUrl" class="red">No URL specified.</p>
		<p v-else-if="repositoryFound" class="green">Repository found.</p>
		<p v-else-if="!repositoryFound" class="red">Repository not found.</p>
	</fieldset>
</template>

<style scoped>
</style>
