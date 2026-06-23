<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
	import { cn } from '$lib/utils/cn';

	let {
		class: className = '',
	}: {
		class?: string;
	} = $props();
</script>

<DropdownMenu.Separator class={cn('bg-line-soft mx-1 my-1 h-px', className)} />
