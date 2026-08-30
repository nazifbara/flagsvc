<script lang="ts" generics="TData">
	import { Input } from '$lib/components/ui/input';
	import { SearchIcon } from '@lucide/svelte';

	type Props<TData> = {
		placeholder?: string;
		getFilterValue: () => string;
		setFilterValue: (value: string) => void;
	};

	const { getFilterValue, setFilterValue, placeholder }: Props<TData> = $props();
</script>

<div class="relative mb-4 flex max-w-sm items-center">
	<span class="pointer-events-none absolute top-1/2 left-2 -translate-y-1/2">
		<SearchIcon class="text-muted-foreground h-4 w-4" />
	</span>
	<Input
		{placeholder}
		value={getFilterValue()}
		class="w-full pl-8"
		onchange={(e) => {
			setFilterValue(e.currentTarget.value);
		}}
		oninput={(e) => {
			setFilterValue(e.currentTarget.value);
		}}
	/>
</div>
