import { computed, ref, watch } from 'vue';

/** Deduplicate the contents of a given array. */
export function deduplicate<T>(array: Array<T>): Array<T> {
	return Array.from(new Set(array));
}

/** Delay the microtask queue by the given duration (ms). */
export async function delay(duration: number): Promise<void> {
	return new Promise(resolve => {
		window.setTimeout(resolve, duration);
	});
}

/** Check if the final segment of the given path is "Tribes Ascend". */
export function pathEndsWithTA(path: string): boolean {
	return /[\\\/]Tribes Ascend[\\\/]?$/.test(path);
}

/** Normalise the given path such that the final segment is "Tribes Ascend". */
export function normalisePath(path: string): string {
	if (pathEndsWithTA(path)) return path;
	else return `${path}\\Tribes Ascend`.replaceAll(/[\\\/]+/g, '\\');
}

/** Get the parent directory of the given path. */
export function parentDirectory(path: string): string {
	const segments = path.replace(/[\\\/]$/, '').split(/[\\\/]/);
	segments.pop();
	return segments.join('\\');
}

/** A reactive semaphore for tracking a group of resource locks. */
export class Semaphore {
	#state = ref(0);

	/** The current value of the semaphore: `true` if any locks exist, `false` if none exist. */
	public readonly value = computed(() => this.#state.value > 0);

	/** Add a lock to the semaphore. */
	public add(): void {
		this.#state.value++;
	}

	/** Remove a lock from the semaphore. */
	public remove(): void {
		if (this.#state.value > 0) {
			this.#state.value--;
		}
	}

	/** Wait for all locks to be resolved. */
	public resolved(): Promise<void> {
		return new Promise(resolve => {
			if (!this.value.value) {
				resolve();
				return;
			}
			const unwatch = watch(this.value, value => {
				if (!value) {
					unwatch();
					resolve();
				}
			});
		});
	}

	/** Create a lock that waits for a given promise to resolve before automatically unlocking. */
	public wait(promise: Promise<any>): void {
		this.add();
		promise.finally(() => this.remove());
	}
}
