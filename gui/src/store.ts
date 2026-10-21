import { getCurrentWindow } from '@tauri-apps/api/window';
import { exists } from '@tauri-apps/plugin-fs';
import { StorageSerializers, useStorage } from '@vueuse/core';
import { createPinia, defineStore } from 'pinia';
import { computed, nextTick, reactive, ref, watch } from 'vue';
import { defaultInstallPath, assetInclusions, models, repository, AppWindow } from './config';
import { BundleSize, Manifest } from './manifest';
import { loadData, saveData } from './persist';
import { Action, DirectoryState, TAI } from './tai';
import { deduplicate, delay, Estimator, formatDuration, normalisePath, parentDirectory, pathEndsWithTA, Semaphore } from './utilities';

const useStore = defineStore('application', () => {
	const isMainWindow = getCurrentWindow().label === AppWindow.Main;
	const action = ref<Action>('install');
	const bundleSize = reactive({
		vanilla: <BundleSize> { packed: 0, unpacked: 0 },
		modded: <BundleSize> { packed: 0, unpacked: 0 },
		custom: <BundleSize> { packed: 0, unpacked: 0 }
	});
	const cancelling = ref(false);
	const createShortcuts = ref(true);
	const estimator = new Estimator();
	const manifest = reactive({
		local: <Manifest | undefined> undefined,
		remote: <Manifest | undefined> undefined
	});
	const model = computed(() => models[step.value]);
	const path = ref(defaultInstallPath);
	const pathState = ref(DirectoryState.Valid);
	const pathVariant = ref('');
	const progressDetail = ref('');
	const progressIndeterminate = ref(false);
	const progressMax = ref(0);
	const progressMeasured = ref(false);
	const progressValue = ref(0);
	const progressComplete = computed(() => progressValue.value >= progressMax.value && progressValue.value);
	const repositoryFound = ref(true);
	const running = ref(false);
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

	function calculateBundleSizes(): void {
		if (!manifest.remote) return;
		const vanillaAssets = manifest.remote.assets.filter(asset => (
			assetInclusions.vanilla.includes(asset.type)
			|| assetInclusions.vanilla.includes(asset.name)
		));
		const customAssets = manifest.remote.assets.filter(asset => (
			requiredCustomAssets.value.has(asset.type)
			|| requiredCustomAssets.value.has(asset.name)
			|| selectedCustomAssets.value.has(asset.type)
			|| selectedCustomAssets.value.has(asset.name)
		));
		bundleSize.vanilla = TAI.determineBundleSize(vanillaAssets);
		bundleSize.modded = TAI.determineBundleSize(manifest.remote.assets);
		bundleSize.custom = TAI.determineBundleSize(customAssets);
	}

	function updateProgressBar(data: string | 0): void {
		if (data === 0) {
			progressIndeterminate.value = false;
			progressDetail.value = '';
			progressMax.value = 0;
			progressValue.value = 0;
			estimator.reset();
			return;
		}
		// A measured download counts bytes, and the assets it counts them for are logged
		// as they finish. Letting both drive the bar would drag it back and forth between
		// the two scales, so the finer measure wins for as long as it is running.
		if (progressMeasured.value) {
			return;
		}
		progressIndeterminate.value = false;
		const result = /\[(\d+)\/(\d+)\]/g.exec(data);
		if (result && result[1] && result[2]) {
			progressMax.value = Number(result[2]);
			progressValue.value = Number(result[1]);
		}
	}

	/** Run the currently selected action through TAI CLI. */
	async function run(): Promise<void> {
		loading.add();
		running.value = true;
		cancelling.value = false;
		progressIndeterminate.value = true;
		const assets = assetInclusions[variant.value];
		if (action.value === 'install' && variant.value === 'custom') {
			assets.push(...requiredCustomAssets.value, ...selectedCustomAssets.value);
		}
		const filter = TAI.filter('include', deduplicate(assets));
		await TAI.run(action.value, path.value, repository, filter, createShortcuts.value);

		// An interrupted install still leaves a partial installation behind, and the selected
		// variant is what identifies it, so record the variant either way.
		if (action.value === 'install') {
			await saveData();
		}

		if (cancelling.value) {
			updateProgressBar(0);
		}
		running.value = false;
		cancelling.value = false;
	}

	/**
	 * Stop the running action.
	 * The CLI winds down and exits by itself, so {@link run} is what finally resolves
	 * the action. This method just requests the cancellation and updates store state.
	 */
	async function cancel(): Promise<void> {
		if (!running.value || cancelling.value) return;
		cancelling.value = true;
		progressDetail.value = '';
		progressIndeterminate.value = true;
		await TAI.cancel();
	}

	function onError(_data: string): void {
		updateProgressBar(0);
	}

	function onInfo(data: string): void {
		updateProgressBar(data);
	}

	function onComplete(_data: string): void {
		// The bar holds an indeterminate state while work it can't measure finishes off.
		// The run is over now, so settle it at full rather than leaving it animating away
		// at nothing. A cancelled run clears the bar for itself.
		if (progressIndeterminate.value && !cancelling.value) {
			progressIndeterminate.value = false;
			progressMax.value = progressMax.value || 1;
			progressValue.value = progressMax.value;
		}
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

	// Refresh the directory state and reload the local manifest whenever the path changes.
	watch(path, refreshDirectory, { immediate: true });
	function refreshDirectory(): void {
		if (!isMainWindow) {
			return;
		}

		nextTick(async () => {
			const lock = loading.add();

			// Reject empty paths.
			if (!path.value) {
				pathState.value = DirectoryState.Invalid;
				lock.remove();
				return;
			}

			// Check that the chosen directory and its parent exist.
			const parent = parentDirectory(path.value);
			const pathExists = await exists(path.value);
			const parentExists = await exists(parent);
			if (!(pathExists || pathEndsWithTA(path.value)) || !parentExists) {
				pathState.value = DirectoryState.Invalid;
				lock.remove();
				return;
			}

			// Check the normalised path with TAI.
			const normalisedPath = normalisePath(path.value);
			let state = await TAI.checkPath(normalisedPath);
			let workingPath = normalisedPath;
	
			// If the normalised path returns valid but not installed,
			// check the actual chosen path for an installation instead.
			if (state === DirectoryState.Valid) {
				state = await TAI.checkPath(path.value);
				workingPath = path.value;
			}

			// If an installation exists, check the type and read the manifest.
			selectedCustomAssets.value.clear();
			if (state === DirectoryState.Installed || state === DirectoryState.NeedsUpdate) {
				const data = await loadData();
				store.pathVariant = data?.selectedVariant ?? '';
				store.createShortcuts = data?.shortcuts ?? true;
				manifest.local = await TAI.getLocalManifest(workingPath);
				for (const asset of manifest.local?.assets ?? []) {
					selectedCustomAssets.value.add(asset.name);
				}
			} else {
				store.pathVariant = '';
			}

			store.pathState = state;
			lock.remove();
		});
	}

	refreshRepository();
	function refreshRepository(): void {
		nextTick(async () => {
			await loading.resolved();
			const lock = loading.add();
			const found = await TAI.checkRepo(repository);
			manifest.remote = await TAI.getRemoteManifest(repository);
			await delay(500);
			repositoryFound.value = found;
			requiredCustomAssets.value = new Set(TAI.determineRequiredAssets(manifest.remote?.assets ?? []).map(asset => asset.name));
			calculateBundleSizes();
			lock.remove();
		});
	}

	// Update the custom bundle size when new assets are selected.
	watch(requiredCustomAssets, calculateBundleSizes);
	watch(selectedCustomAssets, calculateBundleSizes);

	// Follow the CLI's download progress.
	watch(TAI.download, download => {
		progressMeasured.value = Boolean(download);
		if (!download) {
			progressDetail.value = '';
			estimator.reset();
			return;
		}
		// A cancelled run is winding down: leave the bar indeterminate rather than
		// animating out the last few bytes of a download that has been abandoned.
		if (cancelling.value) {
			return;
		}
		progressMax.value = download.size;
		progressValue.value = Math.min(download.bytes, download.size);
		// Every byte has arrived, but the run isn't over: the last assets are still being
		// unpacked, and the prerequisites are installed after them. A bar sitting full and
		// still reads as a run that has finished - or stalled - so it goes back to being
		// indeterminate until the work behind it really is done.
		if (download.bytes >= download.size) {
			progressDetail.value = '';
			progressIndeterminate.value = true;
			return;
		}
		progressIndeterminate.value = false;
		const remaining = estimator.update(download.bytes, download.size);
		progressDetail.value = remaining === undefined ? '' : `${formatDuration(remaining)} left`;
	});

	return {
		action,
		back,
		bundleSize,
		cancel,
		cancelling,
		createShortcuts,
		manifest,
		model,
		next,
		path,
		pathState,
		pathVariant,
		progressComplete,
		progressDetail,
		progressIndeterminate,
		progressMax,
		progressValue,
		refreshDirectory,
		refreshRepository,
		repositoryFound,
		requiredCustomAssets,
		run,
		running,
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
