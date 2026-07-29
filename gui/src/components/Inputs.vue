<script setup lang="ts">
import InstallPathField from './InstallPathField.vue';
import Log from './Log.vue';
import ProgressBar from './ProgressBar.vue';
import VariantSelector from './VariantSelector.vue';
import { store } from '../store';
import { DirectoryState } from '../tai.ts';
</script>

<template>
	<section class="inputs">
		<ol>
			<li
				class="input"
				:class="{ current: store.step === 1 }"
			>
				<VariantSelector
					:disabled="store.loading"
					v-model="store.repo"
				/>
			</li>
			<li
				class="input"
				:class="{ current: store.step === 2 }"
			>
				<InstallPathField
					:disabled="store.loading"
					:refresh="store.refreshPath"
					@update="store.onPathUpdate($event)"
					v-model="store.path"
				/>
			</li>
			<li
				class="input"
				:class="{ current: store.step === 3 }"
			>
				<div class="row">
					<ProgressBar :max="store.progressMax" :value="store.progressValue"/>
					<button
						type="submit"
						:disabled="store.loading || store.pathState === DirectoryState.Invalid || store.isCustomRepo"
						@click="store.run()"
					>{{ store.action }}</button>
				</div>
				<Log />
			</li>
		</ol>
	</section>
</template>

<style scoped>
.inputs {
	display: grid;
	justify-content: stretch;
	align-self: flex-start;
	width: 100%;

	& ol {
		display: flex;
		justify-content: center;
		gap: calc(var(--gap) * 2);
		width: 100%;
		margin: 0;
		padding: 0;
		list-style: none;
	}
}

.input {
	display: none;
	flex-direction: column;
	align-items: center;
	gap: var(--gap);
	width: 100%;
	user-select: none;

	&:where(:focus-visible, :hover) {
		border-color: var(--color-border);
	}

	&.current {
		display: flex;
	}
}
</style>
