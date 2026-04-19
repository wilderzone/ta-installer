<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue';
import { TAI } from '../tai';

const model = defineModel<string>({ required: true });

const repositoryFound = ref(false);
const noUrl = computed(() => !model.value);

async function check(): Promise<void> {
	nextTick(async () => {
		repositoryFound.value = await TAI.checkRepo(model.value);
	});
}

onMounted(() => {
	void check();
});
</script>

<template>
	<fieldset>
		<legend>Repository</legend>
		<input type="text" @change="check()" v-model="model">
		<p v-if="noUrl" class="red">No URL specified.</p>
		<p v-else-if="repositoryFound" class="green">Repository found.</p>
		<p v-else-if="!repositoryFound" class="red">Repository not found.</p>
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
