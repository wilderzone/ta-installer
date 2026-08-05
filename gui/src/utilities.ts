import { computed, ref, watch } from 'vue';

// Convert a given byte count to a string formatted with SI units.
export function byteCountSI(bytes: number): string {
	const unit = 1000;
	const magnitudes = "kMGTPE";

	if (bytes < unit) {
		return `${bytes} B`;
	}

	let divisor = unit;
	let exponent = 0;
	for (let n = bytes / unit; n >= unit; n /= unit) {
		divisor *= unit;
		exponent++;
	}

	return `${(bytes / divisor).toFixed(1)} ${magnitudes[exponent]}B`;
}

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

/**
 * Format a duration (in seconds) as a rough estimate in plain English.
 * A second-accurate countdown on a large download would be false precision,
 * and would spend the whole download flickering.
 */
export function formatDuration(seconds: number): string {
	if (seconds < 10) return 'a few seconds';
	// Rounding can carry a duration up into the next unit, so each threshold is measured
	// against the rounded figure rather than the raw one: "60 minutes" is nobody's idea of
	// an estimate.
	const rounded = Math.round(seconds / 5) * 5;
	if (seconds < 60 && rounded < 60) return `${rounded} seconds`;
	const minutes = Math.round(seconds / 60);
	if (minutes < 2) return '1 minute';
	if (minutes < 60) return `${minutes} minutes`;
	const hours = seconds / 3600;
	if (hours === 1) return `1 hour`
	return `${hours.toFixed(hours < 10 ? 1 : 0)} hours`;
}

/** Check if the final segment of the given path is "Tribes Ascend". */
export function pathEndsWithTA(path: string): boolean {
	return /[\\\/]Tribes Ascend[\\\/]?$/.test(path);
}

/**
 * The path separator this platform uses.
 * Read once at startup: it can't change while the app is running, and every path helper
 * below is synchronous because the components that call them are.
 */
let separator = '/';
export function usePathSeparator(value: string): void {
	separator = value;
}

/** Normalise the given path such that the final segment is "Tribes Ascend". */
export function normalisePath(path: string): string {
	if (pathEndsWithTA(path)) return path;
	// Only the joins are rewritten to the platform separator. Rewriting the whole path
	// would mangle it on Linux, where a forward slash is the separator rather than
	// something to be replaced.
	return `${path}${separator}Tribes Ascend`.replaceAll(/[\\\/]+/g, separator);
}

/** Get the parent directory of the given path. */
export function parentDirectory(path: string): string {
	const trimmed = path.replace(/[\\\/]$/, '');
	const segments = trimmed.split(/[\\\/]/);
	segments.pop();
	const parent = segments.join(separator);
	// A top-level path leaves nothing behind. Its parent is the filesystem root, which
	// exists - returning an empty string would fail the existence check that follows.
	if (!parent) {
		return separator === '/' ? '/' : trimmed;
	}
	return parent;
}

/**
 * A rolling estimate of how much longer a transfer has left to run.
 * The rate is measured across a window of recent samples rather than the whole transfer,
 * so the estimate follows a connection that speeds up or slows down, and each new estimate
 * is eased into the last, so a momentary stall doesn't throw the number around.
 */
export class Estimator {
	/** How much recent history the rate is measured across (ms). */
	private historyWindow = 10000;
	/** How long that history must span before an estimate is worth showing (ms). */
	private settlingPeriod = 1500;
	/** The weight given to each new estimate when easing towards it. */
	private smoothing = 0.2;
	private samples: Array<{ time: number, bytes: number }> = [];
	private estimate: number | undefined;

	/** Record the progress of the transfer, and estimate the time remaining (in seconds). */
	public update(bytes: number, size: number): number | undefined {
		const time = performance.now();
		this.samples.push({ time, bytes });
		while (this.samples.length > 2 && time - this.samples[0].time > this.historyWindow) {
			this.samples.shift();
		}

		// Every byte has arrived. Whatever the transfer is still finishing off, this isn't
		// measuring it, so it has nothing left to estimate.
		if (bytes >= size) {
			this.estimate = undefined;
			return undefined;
		}

		const oldest = this.samples[0];
		const span = time - oldest.time;
		const transferred = bytes - oldest.bytes;
		// Too early to tell, or nothing has moved since the window opened. Holding the
		// previous estimate rides out a stall without the text flickering away.
		if (span < this.settlingPeriod || transferred <= 0) {
			return this.estimate;
		}

		const remaining = (size - bytes) / (transferred / span) / 1000;
		this.estimate = this.estimate === undefined
			? remaining
			: this.estimate + (remaining - this.estimate) * this.smoothing;
		return this.estimate;
	}

	/** Forget the transfer measured so far. */
	public reset(): void {
		this.samples.length = 0;
		this.estimate = undefined;
	}
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
