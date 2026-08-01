import { createRouter, createWebHashHistory } from 'vue-router';
import InstallerView from './views/InstallerView.vue';
import AssetSelectorView from './views/AssetSelectorView.vue';

const routes = [
	{ path: '/', component: InstallerView },
	{ path: '/assetSelect', component: AssetSelectorView },
];

export const router = createRouter({
	history: createWebHashHistory(),
	routes,
});
