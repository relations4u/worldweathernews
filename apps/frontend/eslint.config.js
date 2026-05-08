import js from '@eslint/js';
import ts from 'typescript-eslint';
import svelte from 'eslint-plugin-svelte';
import prettier from 'eslint-config-prettier';
import globals from 'globals';
import svelteParser from 'svelte-eslint-parser';

export default [
	js.configs.recommended,
	...ts.configs.recommended,
	...svelte.configs.recommended,
	prettier,
	...svelte.configs.prettier,
	{
		languageOptions: {
			globals: {
				...globals.browser,
				...globals.node
			}
		}
	},
	{
		files: ['**/*.svelte', '**/*.svelte.ts', '**/*.svelte.js'],
		languageOptions: {
			parser: svelteParser,
			parserOptions: {
				parser: ts.parser,
				svelteConfig: './svelte.config.js'
			}
		}
	},
	{
		ignores: [
			'build/',
			'.svelte-kit/',
			'dist/',
			'node_modules/',
			'package/',
			'coverage/',
			'pnpm-lock.yaml',
			// Generated von packages/api-schema (openapi-typescript).
			'src/lib/api/types.gen.ts',
			// Paraglide-generierte i18n-Files (kompiliert aus messages/ + project.inlang/).
			'src/lib/paraglide/'
		]
	}
];
