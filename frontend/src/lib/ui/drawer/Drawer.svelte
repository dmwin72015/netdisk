<script lang="ts">
	import type { Snippet } from 'svelte';
	import * as DrawerBase from './base';

	type Side = 'left' | 'right' | 'top' | 'bottom';

	let {
		open = $bindable(false),
		onOpenChange,
		side = 'right',
		title,
		description,
		children,
		trigger,
		class: className = '',
	}: {
		open?: boolean;
		onOpenChange?: (open: boolean) => void;
		side?: Side;
		title?: string;
		description?: string;
		children: Snippet;
		trigger?: Snippet;
		class?: string;
	} = $props();
</script>

<DrawerBase.Root bind:open {onOpenChange}>
	{#if trigger}
		<DrawerBase.Trigger>
			{@render trigger()}
		</DrawerBase.Trigger>
	{/if}

	<DrawerBase.Overlay />

	<DrawerBase.Content {side} class={className}>
		{#if title || description}
			<DrawerBase.Header>
				{#if title}
					<DrawerBase.Title>{title}</DrawerBase.Title>
				{/if}
				{#if description}
					<DrawerBase.Description>{description}</DrawerBase.Description>
				{/if}
			</DrawerBase.Header>
		{/if}

		{@render children()}

		<DrawerBase.Close
			class="absolute right-3 top-3"
		/>
	</DrawerBase.Content>
</DrawerBase.Root>
