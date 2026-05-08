import adapter from '@sveltejs/adapter-node';
import { mdsvex } from 'mdsvex';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	extensions: ['.svelte', '.svx', '.md'],
	compilerOptions: {
		// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
		runes: ({ filename }) => (filename.split(/[/\\]/).includes('node_modules') ? undefined : true)
	},
	preprocess: [
		mdsvex({
			extensions: ['.svx', '.md']
		})
	],
	kit: {
		adapter: adapter({
			out: 'build',
			precompress: true,
			envPrefix: 'WWN_FRONTEND_'
		})
	}
};

export default config;
