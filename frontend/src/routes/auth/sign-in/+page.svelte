<script lang="ts">
	import { superForm } from 'sveltekit-superforms';
	import { zod4Client } from 'sveltekit-superforms/adapters';
	import * as Form from '$lib/components/ui/form';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import Separator from '$lib/components/ui/separator/separator.svelte';
	import { z } from 'zod';
	import Spinner from '$lib/components/ui/spinner/spinner.svelte';

	let { data } = $props();

	const signInSchema = z.object({
		email: z.email('Invalid email address'),
		password: z.string().min(1, 'Password is required')
	});

	const form = superForm(data.form, {
		validators: zod4Client(signInSchema),
		taintedMessage: 'You have unsaved changes'
	});

	const { form: formData, enhance, message, submitting } = form;
</script>

<div class="bg-background flex min-h-screen items-center justify-center px-4 py-12">
	<Card.Root class="w-full max-w-md">
		<Card.Header>
			<Card.Title class="text-2xl">Sign In</Card.Title>
			<Card.Description>Sign in to your 1on1 Coaching account</Card.Description>
		</Card.Header>
		<Card.Content>
			<form method="POST" use:enhance>
				<div class="space-y-4">
					<!-- Email Field -->
					<Form.Field {form} name="email" class="space-y-2">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>Email Address</Form.Label>
								<Input
									{...props}
									type="email"
									placeholder="you@example.com"
									bind:value={$formData.email}
									disabled={$submitting}
								/>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>

					<!-- Password Field -->
					<Form.Field {form} name="password" class="space-y-2">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>Password</Form.Label>
								<Input
									{...props}
									type="password"
									placeholder="••••••••"
									bind:value={$formData.password}
									disabled={$submitting}
								/>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>

					<!-- Error Message -->
					{#if $message}
						<div class="bg-destructive/10 text-destructive rounded-md p-3 text-sm">
							{$message}
						</div>
					{/if}

					<!-- Submit Button -->
					<Button type="submit" class="w-full" disabled={$submitting}>
						{#if $submitting}
							<Spinner />
						{/if}
						Sign in
					</Button>
				</div>
			</form>

			<!-- Divider -->
			<div class="my-6 flex items-center gap-2">
				<Separator class="flex-1" />
				<span class="text-muted-foreground text-xs">or</span>
				<Separator class="flex-1" />
			</div>

			<!-- Sign Up Link -->
			<div class="text-muted-foreground text-center text-sm">
				Don't have an account?
				<a href="/auth/sign-up" class="text-primary font-semibold hover:underline">Create one</a>
			</div>
		</Card.Content>
	</Card.Root>
</div>
