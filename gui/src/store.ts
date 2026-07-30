import { createPinia, defineStore } from 'pinia';
import { computed, ref, watch } from 'vue';
import { defaultInstallPath, filters, repository } from './config';
import { Action, DirectoryState, TAI } from './tai';
import { saveData } from './persist';

const useStore = defineStore('application', () => {
	const models = ['GravCycle', 'GravCycle', 'Shrike', 'Beowulf'];
	const action = ref<Action>('install');
	const loading = ref(false);
	const path = ref(defaultInstallPath);
	const pathState = ref(DirectoryState.Valid);
	const pathVariant = ref('');
	const progressIndeterminate = ref(false);
	const progressMax = ref(0);
	const progressValue = ref(0);
	const refreshPath = ref(1);
	const step = ref(1);
	const step1Error = ref(false);
	const step2Error = computed(() => pathState.value === DirectoryState.Invalid);
	const step3Error = ref(false);
	const variant = ref('modded');

	const isCustomVariant = computed(() => variant.value === 'custom');
	const model = computed(() => models[step.value]);

	function back(): void {
		if (step.value - 1 < 1) return;
		step.value--;
	}

	function next(): void {
		if (step.value + 1 > 3) return;
		step.value++;
	}

	function updateProgressBar(data: string | 0): void {
		progressIndeterminate.value = false;
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

	async function run(): Promise<void> {
		loading.value = true;
		progressIndeterminate.value = true;
		await TAI.run(action.value, path.value, repository, filters[variant.value]);
		if (action.value === 'install') {
			await saveData();
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

	TAI.attach(onInfo, onError, onComplete);

	// Update the action when the path state changes.
	watch(pathState, value => {
		switch (value) {
			case DirectoryState.Installed:
				if (action.value === 'install') {
					action.value = 'verify';
				}
				break;
			case DirectoryState.Valid:
				action.value = 'install';
				break;
		}
	});

	return {
		action,
		back,
		loading,
		isCustomVariant,
		model,
		path,
		pathState,
		pathVariant,
		progressIndeterminate,
		progressMax,
		progressValue,
		refreshPath,
		next,
		run,
		step,
		step1Error,
		step2Error,
		step3Error,
		updateProgressBar,
		variant
	};
});

export const pinia = createPinia();
export const store = useStore(pinia);
