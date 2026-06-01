import adapter from '@sveltejs/adapter-static';
import {vitePreprocess} from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		// adapter-auto only supports some environments, see https://kit.svelte.dev/docs/adapter-auto for a list.
		// If your environment is not supported, or you settled on a specific environment, switch out the adapter.
		// See https://kit.svelte.dev/docs/adapters for more information about adapters.
		adapter: adapter({
			pages: 'dist',
			assets: 'dist',
			fallback: undefined,
			precompress: false,
			strict: true
		}),
		alias: {
			'$bindings': './bindings',
			'$wails': './bindings/github.com/wailsapp/wails/v3/internal',
			'$hindsight': './bindings/github.com/zerebos/hindsight',
			'$app': './bindings/github.com/zerebos/hindsight/internal/app',
			'$db': './bindings/github.com/zerebos/hindsight/internal/db',
			'$dbgen': './bindings/github.com/zerebos/hindsight/internal/db/generated',
			'$browser': './bindings/github.com/zerebos/hindsight/internal/browser',
			'$ingestion': './bindings/github.com/zerebos/hindsight/internal/ingestion',
		}
	}
};

export default config;
