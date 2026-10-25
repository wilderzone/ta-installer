<script setup lang="ts">
import { store } from '../store.ts';
import InstallPathField from './InstallPathField.vue';
import Log from './Log.vue';
import ProgressBar from './ProgressBar.vue';
import StepNavigation from './StepNavigation.vue';
import Steps from './Steps.vue';
import VariantSelector from './VariantSelector.vue';
</script>

<template>
	<form
		class="configurator"
		aria-label="Installation configuration"
		action=""
		@submit.prevent
	>
		<Steps v-grid:steps />
		<ol class="panels" :style="{ '--step': store.step }">
			<li
				role="tabpanel"
				id="panel-step-1"
				aria-labelledby="tab-step-1"
				class="panel"
				:inert="store.step !== 1"
			>
				<VariantSelector />
			</li>
			<li
				role="tabpanel"
				id="panel-step-2"
				aria-labelledby="tab-step-2"
				class="panel"
				:inert="store.step !== 2"
			>
				<InstallPathField />
			</li>
			<li
				role="tabpanel"
				id="panel-step-3"
				aria-labelledby="tab-step-3"
				class="panel"
				:inert="store.step !== 3"
			>
				<div class="output">
					<ProgressBar
						:aria-label="`${store.action} progress`"
						:detail="store.progressDetail"
						:indeterminate="store.progressIndeterminate"
						:max="store.progressMax"
						:value="store.progressValue"
					/>
					<Log />
				</div>
			</li>
		</ol>
		<StepNavigation v-grid:navigation />
	</form>
</template>

<style scoped>
.configurator {
	display: grid;
	grid-template:
		'steps     ' auto
		'panels    ' 1fr
		'navigation' auto
		/ 1fr;
	gap: var(--gap-m);
	height: 100%;
}

.panels {
	--panel-gap: calc(var(--gap-m) * 2);
	grid-area: panels;
	align-self: flex-start;
	display: grid;
	grid-auto-flow: column;
	grid-auto-columns: 100%;
	grid-template-rows: 100%;
	gap: var(--panel-gap);
	width: 100%;
	height: 100%;
	min-width: 0;
	min-height: 0;
	margin: 0;
	padding: 0;
	list-style: none;
	overflow-x: clip;

	.panel {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--gap-m);
		width: 100%;
		height: 100%;
		min-height: 0;
		transform: translateX(calc((1 - var(--step)) * (100% + var(--panel-gap))));
		transition: 0.3s ease transform, 0.4s ease opacity;

		&:where(:focus-visible, :hover) {
			border-color: var(--color-border);
		}

		&[inert] {
			opacity: 0;
		}

		@media (prefers-reduced-motion: reduce) {
			transition: 0.4s ease opacity;
		}
	}
}

.output {
	display: grid;
	grid-template-rows: auto 1fr;
	width: 100%;
	height: 100%;
}
</style>
