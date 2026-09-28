import { defineConfig } from 'vitepress';

// https://vitepress.dev/reference/site-config
export default defineConfig({
	title: "TAI Documentation",
	description: "Documentation for the Tribes: Ascend installer.",
	appearance: 'force-dark',
	lastUpdated: true,
	themeConfig: {
		// https://vitepress.dev/reference/default-theme-config

		logo: '/assets/logo.png',

		sidebar: [
		],

		socialLinks: [
			{ icon: 'github', link: 'https://github.com/wilderzone/ta-installer' }
		],

		footer: {
			message: [
				{ text: 'Ask a question', link: 'https://discord.gg/dd8JgzJ' },
				{ text: 'Give feedback', link: 'https://feedback.wilderzone.org/' },
				{ text: 'Donate', link: 'https://ko-fi.com/wilderzone' },
				{ text: 'Source code', link: 'https://github.com/wilderzone/ta-installer' }
			].map(item => `<a href="${item.link}" target="_blank" rel="noreferrer">${item.text}</a>`).join(' '),
			copyright: '© 2026 Wilderzone.org'
		}
	}
});
