import { Directive } from 'vue';

declare module 'vue' {
	interface GlobalDirectives {
		/** Attach the element to a given CSS grid area. */
		vGrid: typeof vGrid;
	}
}

/** Attach the element to a given CSS grid area. */
export const vGrid: Directive<HTMLElement, never> = {
	beforeMount(element, binding) {
		if (binding.arg) {
			element.style.gridArea = binding.arg;
		}
	},
	updated(element, binding) {
		if (binding.arg) {
			element.style.gridArea = binding.arg;
		}
	}
};
