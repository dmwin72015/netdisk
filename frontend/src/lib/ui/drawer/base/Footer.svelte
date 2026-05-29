<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		children,
		class: className = '',
	}: {
		children: Snippet;
		class?: string;
	} = $props();
</script>

<div class="flex flex-col-reverse gap-2 p-4 pt-2 sm:flex-row sm:justify-end {className}">
	{@render children()}
</div>
