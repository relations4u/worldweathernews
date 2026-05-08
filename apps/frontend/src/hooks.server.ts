import type { Handle } from '@sveltejs/kit';
import { paraglideMiddleware } from '$lib/paraglide/server';

export const handle: Handle = ({ event, resolve }) => {
	return paraglideMiddleware(event.request, ({ request, locale }) => {
		event.request = request;
		event.locals.locale = locale;
		return resolve(event, {
			transformPageChunk: ({ html }) => html.replace('%paraglide.lang%', locale)
		});
	});
};
