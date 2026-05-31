import type { LayoutLoad } from './$types';

export const load: LayoutLoad = ({ url }) => {
	return {
		isAuthPage: url.pathname === '/login' || url.pathname === '/register',
	};
};
