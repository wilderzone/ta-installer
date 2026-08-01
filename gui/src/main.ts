import { createApp } from 'vue';
import './root.css';
import App from './App.vue';
import { vGrid } from './directives';
import { router } from './router.ts';
import { handleShortcuts } from './shortcuts';
import { pinia } from './store.ts';

createApp(App)
	.use(router)
	.use(pinia)
	.directive('grid', vGrid)
	.mount('#app');

handleShortcuts();
