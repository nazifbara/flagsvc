import { fail, redirect } from '@sveltejs/kit';
import { message, superValidate } from 'sveltekit-superforms';
import { zod4 } from 'sveltekit-superforms/adapters';
import { z } from 'zod';

const signInSchema = z.object({
	email: z.email('Invalid email address'),
	password: z.string().min(1, 'Password is required')
});

export async function load({ locals }) {
	const { userId } = locals;

	if (userId) {
		throw redirect(303, '/app');
	}

	const form = await superValidate(zod4(signInSchema));
	return { form };
}

export const actions = {
	default: async ({ request, fetch }) => {
		const form = await superValidate(request, zod4(signInSchema));

		if (!form.valid) {
			return fail(400, { form });
		}

		try {
			// Call the Convex login mutation
			const response = await fetch('/api/auth/sign-in', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					email: form.data.email,
					password: form.data.password
				})
			});

			if (!response.ok) {
				const error = await response.json();
				return message(form, error.message || 'Sign-in failed', { status: 401 });
			}
		} catch (error) {
			if (error instanceof Error) {
				if (error.message.includes('redirect')) throw error;
				return fail(500, { form, error: error.message });
			}
			return fail(500, { form, error: 'An unexpected error occurred' });
		}

		redirect(303, '/app');
	}
};
