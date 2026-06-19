import type { Handle, HandleServerError } from '@sveltejs/kit';
import { paraglideMiddleware } from '$lib/paraglide/server';
import { getTextDirection } from '$lib/paraglide/runtime';

const paraglideHandle: Handle = ({ event, resolve }) =>
	paraglideMiddleware(event.request, ({ request: localizedRequest, locale }) => {
		event.request = localizedRequest;
		return resolve(event, {
			transformPageChunk: ({ html }) => {
				return html
					.replace('%lang%', locale)
					.replace('%dir%', getTextDirection(locale));
			}
		});
	});

export const handle: Handle = paraglideHandle;

export const handleError: HandleServerError = ({ error, status, message }) => {
	console.error(`[SSR Error] status=${status} message=${message}`);
	if (error instanceof Error) {
		console.error('[SSR Error stack]', error.stack);
	}
	return { message: 'Internal Error' };
};
