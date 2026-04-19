import { createApp } from 'vue';
import './root.css';
import App from './App.vue';
import { handleShortcuts } from './shortcuts';

createApp(App).mount('#app');
handleShortcuts();
