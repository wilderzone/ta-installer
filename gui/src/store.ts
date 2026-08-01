import { StorageSerializers, useStorage } from '@vueuse/core';
import { createPinia, defineStore } from 'pinia';
import { computed, nextTick, reactive, ref, watch } from 'vue';
import { defaultInstallPath, assetInclusions, models, repository } from './config';
import { Manifest } from './manifest';
import { loadData, saveData } from './persist';
import { Action, DirectoryState, TAI } from './tai';
import { deduplicate, delay, normalisePath, parentDirectory, pathEndsWithTA, Semaphore } from './utilities';
import { exists } from '@tauri-apps/plugin-fs';

const useStore = defineStore('application', () => {
	const action = ref<Action>('install');
	const manifest = reactive({
		local: <Manifest | undefined> undefined,
		remote: <Manifest | undefined> undefined
	});
	const model = computed(() => models[step.value]);
	const path = ref(defaultInstallPath);
	const pathState = ref(DirectoryState.Valid);
	const pathVariant = ref('');
	const progressIndeterminate = ref(false);
	const progressMax = ref(0);
	const progressValue = ref(0);
	const progressComplete = computed(() => progressValue.value >= progressMax.value && progressValue.value);
	const repositoryFound = ref(true);
	const requiredCustomAssets = useStorage<Set<string>>('requiredCustomAssets', new Set(), localStorage, { serializer: StorageSerializers.set });
	const selectedCustomAssets = useStorage<Set<string>>('selectedCustomAssets', new Set(), localStorage, { serializer: StorageSerializers.set });
	const step = ref(1);
	const step1Error = computed(() => !repositoryFound.value);
	const step2Error = computed(() => pathState.value === DirectoryState.Invalid);
	const step3Error = ref(false);
	const variant = ref('modded');

	/** Return to the previous installation step. */
	function back(): void {
		if (step.value - 1 < 1) return;
		step.value--;
	}

	/** Proceed to the next installation step. */
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

	/** Run the currently selected action through TAI CLI. */
	async function run(): Promise<void> {
		loading.add();
		progressIndeterminate.value = true;
		const assets = assetInclusions[variant.value];
		if (action.value === 'install' && variant.value === 'custom') {
			assets.push(...requiredCustomAssets.value, ...selectedCustomAssets.value);
		}
		const filter = TAI.filter('include', deduplicate(assets));
		await TAI.run(action.value, path.value, repository, filter);
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
		refreshDirectory();
		loading.remove();
	}

	TAI.attach(onInfo, onError, onComplete);

	// Update the action when the path state changes.
	watch(pathState, value => {
		switch (value) {
			case DirectoryState.Installed:
				if (action.value === 'install' || action.value === 'update') {
					action.value = 'verify';
				}
				break;
			case DirectoryState.NeedsUpdate:
				action.value = 'update';
				break;
			case DirectoryState.Valid:
				action.value = 'install';
				break;
		}
	});

	// Reload the local manifest whenever the path changes.
	watch(path, refreshDirectory, { immediate: true });
	function refreshDirectory(): void {
		nextTick(async () => {
			loading.add();
			const parent = parentDirectory(path.value);

			// Reject empty paths.
			if (!path.value) {
				pathState.value = DirectoryState.Invalid;
				loading.remove();
				return;
			}

			// Check that the chosen directory and its parent exist.
			const pathExists = await exists(path.value);
			const parentExists = await exists(parent);
			if (!(pathExists || pathEndsWithTA(path.value)) || !parentExists) {
				pathState.value = DirectoryState.Invalid;
				loading.remove();
				return;
			}

			// Check the normalised path with TAI.
			const normalisedPath = normalisePath(path.value);
			let state = await TAI.checkPath(normalisedPath);
	
			// If the normalised path returns valid but not installed,
			// check the actual chosen path for an installation instead.
			if (state === DirectoryState.Valid) {
				state = await TAI.checkPath(path.value);
			}

			// If an installation exists, check the type.
			if (state === DirectoryState.Installed || state === DirectoryState.NeedsUpdate) {
				const data = await loadData();
				store.pathVariant = data?.selectedVariant ?? '';
			} else {
				store.pathVariant = '';
			}

			store.pathState = state;
			loading.remove();
		});
	}

	refreshRepository();
	function refreshRepository(): void {
		nextTick(async () => {
			await loading.resolved();
			loading.add();
			const found = await TAI.checkRepo(repository);
			manifest.remote = await TAI.getRemoteManifest(repository);
			await delay(500);
			repositoryFound.value = found;
			requiredCustomAssets.value = new Set(TAI.determineRequiredAssets(manifest.remote?.assets ?? []).map(asset => asset.name));
			loading.remove();
		});
	}

	return {
		action,
		back,
		manifest,
		model,
		next,
		path,
		pathState,
		pathVariant,
		progressComplete,
		progressIndeterminate,
		progressMax,
		progressValue,
		repositoryFound,
		requiredCustomAssets,
		run,
		selectedCustomAssets,
		step,
		step1Error,
		step2Error,
		step3Error,
		updateProgressBar,
		variant
	};
});


export const loading = new Semaphore();
export const pinia = createPinia();
export const store = useStore(pinia);
