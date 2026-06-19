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

export const handle: Handle = async ({ event, resolve }) => {
	try {
		return await paraglideHandle({ event, resolve });
	} catch (e) {
		process.stderr.write(`[SSR Handle Fatal] ${e}\n`);
		if (e instanceof Error) {
			process.stderr.write(`[SSR Handle Fatal stack] ${e.stack}\n`);
		}
		throw e;
	}
};

export const handleError: HandleServerError = ({ error, status, message }) => {
	process.stderr.write(`[SSR Error] status=${status} message=${message}\n`);
	if (error instanceof Error) {
		process.stderr.write(`[SSR Error stack] ${error.stack}\n`);
	}
	return { message: 'Internal Error' };
};
