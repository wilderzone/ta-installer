<script setup lang="ts">
import { computed } from 'vue';
import { Asset } from '../manifest';
import { store } from '../store';

const categories = computed(() => {
	const output: Record<string, Array<Asset>> = {};
	const assets = store.manifest.remote?.assets ?? [];
	for (const asset of assets) {
		if (asset.type in output) {
			output[asset.type].push(asset);
		} else {
			output[asset.type] = [asset];
		}
	}
	return output;
});

function assetID(asset: Asset): string {
	return `asset-${asset.name.replaceAll(/\s/g, '_')}`;
}

function toggle(event: InputEvent, asset: Asset): void {
	if (!(event.target instanceof HTMLInputElement)) {
		return;
	}
	if (store.requiredCustomAssets.has(asset.name)) {
		return;
	}

	if (event.target.checked) {
		store.selectedCustomAssets.add(asset.name);
	} else {
		store.selectedCustomAssets.delete(asset.name);
	}
}
</script>

<template>
	<ul class="asset-list">
		<li v-for="assets, category of categories" class="category">
			<p>{{ category }}</p>
			<ul>
				<li v-for="asset of assets" class="asset">
					<input
						type="checkbox"
						:name="assetID(asset)"
						:id="assetID(asset)"
						:checked="store.requiredCustomAssets.has(asset.name) || store.selectedCustomAssets.has(asset.name)"
						:disabled="store.requiredCustomAssets.has(asset.name)"
						@input="toggle($event, asset)"
					>
					<label :for="assetID(asset)">{{ asset.name }}</label>
				</li>
			</ul>
		</li>
	</ul>
</template>

<style scoped>
.asset-list {
	display: flex;
	flex-direction: column;
	gap: var(--gap-s);
	height: 100%;
	margin: 0;
	padding: var(--gap-s) var(--gap-s) 0;
	list-style: none;
	overflow: auto;
}

.category {
	display: flex;
	flex-direction: column;
	gap: var(--gap-xs);
	padding: 0 0 var(--gap-s);
	background: var(--color-card);
	border-radius: var(--border-radius);

	& > p {
		margin: 0;
		padding: var(--gap-xs) 0;
		font-style: italic;
		text-transform: uppercase;
		background: var(--color-card);
		border-radius: inherit;

		&::before {
			content: '////////';
			margin-inline-end: 1ch;
			mask-image: linear-gradient(to left, #0004, #0000);
		}

		&::after {
			content: '////////////////////////////////////////////////////////////////////////////////';
			margin-inline-start: 1ch;
			mask-image: linear-gradient(to right, #0004, #0000);
		}
	}

	& ul {
		margin: 0;
		padding: 0 var(--gap-s);
		list-style: none;
	}
}

.asset {
	display: flex;
	align-items: center;
	gap: 1ch;

	& label {
		width: 100%;
	}

	& input[type="checkbox"]:focus-visible {
		outline: var(--focus-outline);
	}
}
</style>
