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

<div class="flex flex-col space-y-1.5 {className}">
	{@render children()}
</div>
