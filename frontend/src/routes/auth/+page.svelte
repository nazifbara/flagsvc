<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import Separator from '$lib/components/ui/separator/separator.svelte';
</script>

<div class="bg-background flex min-h-screen items-center justify-center px-4 py-12">
	<Card.Root class="w-full max-w-md">
		<Card.Header>
			<Card.Title class="text-2xl">Welcome Back</Card.Title>
			<Card.Description>Sign in to your 1on1 Coaching account</Card.Description>
		</Card.Header>
		<Card.Content class="flex flex-col gap-4">
			<Button variant="outline">
				<a href="/auth/sign-in">Sign in</a>
			</Button>
			<div
				class="text-muted-foreground grid w-full grid-cols-[1fr_auto_1fr] items-center gap-2 text-sm"
			>
				<Separator /> or <Separator />
			</div>
			<Button>
				<a href="/auth/sign-up">Create account</a>
			</Button>
		</Card.Content>
	</Card.Root>
</div>
