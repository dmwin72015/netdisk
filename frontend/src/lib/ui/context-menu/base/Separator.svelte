<script lang="ts">
	import { ContextMenu } from 'bits-ui';
	import { cn } from '$lib/utils/cn';

	let {
		class: className = '',
	}: {
		class?: string;
	} = $props();
</script>

<ContextMenu.Separator class={cn('bg-line-soft mx-1 my-1 h-px', className)} />
