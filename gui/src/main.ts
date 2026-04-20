import { createApp } from 'vue';
import './root.css';
import App from './App.vue';
import { handleShortcuts } from './shortcuts';
import { extractSidecar } from './sidecar';

createApp(App).mount('#app');
handleShortcuts();
extractSidecar();
