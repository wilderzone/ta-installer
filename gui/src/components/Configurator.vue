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
	<form class="configurator" action="" @submit.prevent>
		<Steps v-grid:steps />
		<ol class="panels">
			<li
				class="panel"
				:class="{ current: store.step === 1 }"
			>
				<VariantSelector />
			</li>
			<li
				class="panel"
				:class="{ current: store.step === 2 }"
			>
				<InstallPathField />
			</li>
			<li
				class="panel"
				:class="{ current: store.step === 3 }"
			>
				<div class="output">
					<ProgressBar
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
	grid-area: panels;
	display: flex;
	justify-content: center;
	align-self: flex-start;
	gap: calc(var(--gap-m) * 2);
	width: 100%;
	height: 100%;
	min-height: 0;
	margin: 0;
	padding: 0;
	list-style: none;

	.panel {
		display: none;
		flex-direction: column;
		align-items: center;
		gap: var(--gap-m);
		width: 100%;
		height: 100%;

		&:where(:focus-visible, :hover) {
			border-color: var(--color-border);
		}

		&.current {
			display: flex;
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
