import { redirect } from '@sveltejs/kit';

export const load = async ({ locals: { userId } }) => {
	if (userId) {
		throw redirect(303, '/app');
	}

	return {};
};
