<script lang="ts">
	import { DropdownMenu } from 'bits-ui';

	let {
		class: className = '',
	}: {
		class?: string;
	} = $props();
</script>

<DropdownMenu.Separator class="mx-2 my-1 h-px bg-gray-100 {className}" />
