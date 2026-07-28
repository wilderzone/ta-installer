import { createPinia, defineStore } from 'pinia';
import { computed, ref } from 'vue';

const useStore = defineStore('loadouts', () => {
	const model = computed(() => {
		const models = ['GravCycle', 'GravCycle', 'Shrike', 'Beowulf'];
		return models[step.value];
	});

	return {
		model,
	};
});

export const pinia = createPinia();
export const store = useStore(pinia);
