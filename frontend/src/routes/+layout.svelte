<script lang="ts">
	import '../app.css';
	import { ModeWatcher } from 'mode-watcher';
	import { Toaster } from '$lib/components/ui/sonner';

	const { children } = $props();
</script>

<Toaster />
<ModeWatcher />

<main class="mx-auto max-w-7xl px-4 2xl:px-0">
	{@render children?.()}
</main>
