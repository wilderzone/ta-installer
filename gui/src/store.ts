import { createPinia, defineStore } from 'pinia';
import { computed, ref } from 'vue';
import { repository } from './config';
import { Action, DirectoryState, TAI } from './tai';

const useStore = defineStore('loadouts', () => {
	const models = ['GravCycle', 'GravCycle', 'Shrike', 'Beowulf'];
	const action = ref<Action>('install');
	const loading = ref(false);
	const path = ref('');
	const pathState = ref(DirectoryState.Invalid);
	const progressMax = ref(0);
	const progressValue = ref(0);
	const refreshPath = ref(1);
	const repo = ref('vanilla');
	const step = ref(1);
	const step1Error = ref(false);
	const step2Error = ref(false);
	const step3Error = ref(false);

	const isCustomRepo = computed(() => repo.value === 'custom');
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

	return {
		action,
		back,
		loading,
		isCustomRepo,
		model,
		onPathUpdate,
		path,
		pathState,
		progressMax,
		progressValue,
		refreshPath,
		repo,
		next,
		run,
		step,
		step1Error,
		step2Error,
		step3Error,
		updateProgressBar
	};
});

export const pinia = createPinia();
export const store = useStore(pinia);
