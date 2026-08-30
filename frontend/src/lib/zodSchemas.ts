import { z } from 'zod';

export const signUpSchema = z
	.object({
		email: z.email('Invalid email address'),
		firstName: z.string().min(2, 'Name must be at least 2 characters'),
		lastName: z.string().min(2, 'Name must be at least 2 characters'),
		password: z.string().min(8, 'Password must be at least 8 characters'),
		timezone: z.string(),
		confirmPassword: z.string()
	})
	.refine((data) => data.password === data.confirmPassword, {
		message: "Passwords don't match",
		path: ['confirmPassword']
	});
