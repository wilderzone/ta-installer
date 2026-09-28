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
			{
				text: 'Overview',
				link: '/',
				items: [
					{ text: 'Glossary', link: '/#glossary' }
				]
			},
			{
				text: 'CLI',
				base: '/cli',
				link: '/',
				items: [
					{ text: 'Subcommands', link: '/subcommands' },
					{ text: 'Options', link: '/options' },
					{ text: 'Output', link: '/output' },
					{ text: 'Dependency Resolution', link: '/dependency-resolution' }
				]
			},
			{
				text: 'GUI',
				base: '/gui',
				link: '/',
				items: [
					{ text: 'Game Type', link: '/game-type' },
					{ text: 'Location', link: '/location' },
					{ text: 'Action', link: '/action' }
				]
			},
			{
				text: 'Development',
				base: '/development',
				link: '/',
				items: [
					{ text: 'Project Structure', link: '/project-structure' },
					{ text: 'Running on Linux', link: '/running-on-linux' },
					{ text: 'Installation Utilities', link: '/installation-utilities' },
					{ text: 'Building', link: '/building' }
				]
			}
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
