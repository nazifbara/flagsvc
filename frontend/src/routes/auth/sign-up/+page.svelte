<script lang="ts">
	import { defaults, setMessage, superForm } from 'sveltekit-superforms';
	import { zod4, zod4Client } from 'sveltekit-superforms/adapters';
	import * as Form from '$lib/components/ui/form';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import Separator from '$lib/components/ui/separator/separator.svelte';
	import { Input } from '$lib/components/ui/input';
	import { signUpSchema } from '$lib/zodSchemas';
	import Spinner from '$lib/components/ui/spinner/spinner.svelte';
	import { authClient } from '$lib/client';
	import { ConnectError } from '@connectrpc/connect';
	import { goto } from '$app/navigation';

	let isLoading = $state(false);

	const form = superForm(defaults(zod4(signUpSchema)), {
		SPA: true,
		validators: zod4Client(signUpSchema),
		invalidateAll: false,
		taintedMessage: 'You have unsaved changes',
		resetForm: false,
		onUpdate: async ({ form }) => {
			if (!form.valid) return;
			try {
				isLoading = true;
				await authClient.signUp({ ...form.data });
				goto('/auth/sign-in');
			} catch (err) {
				if (err instanceof ConnectError) {
					if (err.code === 6) {
						setMessage(form, 'Account already exist');
					}
					console.log({ err });
				} else {
					setMessage(form, 'Something went wrong');
				}
			} finally {
				isLoading = false;
			}
		}
	});

	const { form: formData, enhance, submitting, message } = form;
</script>

<div class="bg-background flex min-h-screen items-center justify-center px-4 py-12">
	<Card.Root class="w-full max-w-md">
		<Card.Header>
			<Card.Title class="text-2xl">Create Account</Card.Title>
			<Card.Description>Sign up to get started with 1on1 Coaching</Card.Description>
		</Card.Header>
		<Card.Content>
			<form method="POST" use:enhance>
				<div class="space-y-4">
					<!-- Name Field -->
					<Form.Field {form} name="firstName" class="space-y-2">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>First name</Form.Label>
								<Input
									{...props}
									placeholder="John"
									bind:value={$formData.firstName}
									disabled={isLoading}
								/>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>

					<!-- Name Field -->
					<Form.Field {form} name="lastName" class="space-y-2">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>Last name</Form.Label>
								<Input
									{...props}
									placeholder="Doe"
									bind:value={$formData.lastName}
									disabled={isLoading}
								/>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>
					<!-- Email Field -->
					<Form.Field {form} name="email" class="space-y-2">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>Email Address</Form.Label>
								<Input
									{...props}
									placeholder="you@example.com"
									bind:value={$formData.email}
									disabled={isLoading}
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
									disabled={isLoading}
								/>
							{/snippet}
						</Form.Control>
						<Form.Description>At least 8 characters</Form.Description>
						<Form.FieldErrors />
					</Form.Field>

					<!-- Confirm Password Field -->
					<Form.Field {form} name="confirmPassword" class="space-y-2">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>Confirm Password</Form.Label>
								<Input
									{...props}
									type="password"
									placeholder="••••••••"
									bind:value={$formData.confirmPassword}
									disabled={isLoading}
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
					<Button type="submit" class="w-full" disabled={isLoading}>
						{#if $submitting}
							<Spinner />
						{/if}
						Create account
					</Button>
				</div>
			</form>

			<!-- Divider -->
			<div class="my-6 flex items-center gap-2">
				<Separator class="flex-1" />
				<span class="text-muted-foreground text-xs">or</span>
				<Separator class="flex-1" />
			</div>

			<!-- Sign In Link -->
			<div class="text-muted-foreground text-center text-sm">
				Already have an account?
				<a href="/auth/sign-in" class="text-primary font-semibold hover:underline">Sign in</a>
			</div>
		</Card.Content>
	</Card.Root>
</div>
