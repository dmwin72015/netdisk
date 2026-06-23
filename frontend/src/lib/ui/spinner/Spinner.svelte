<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils/cn';

	let {
		size = 24,
		color = 'currentColor',
		strokeWidth = 2,
		class: className = '',
		...rest
	}: {
		size?: number;
		color?: string;
		strokeWidth?: number;
		class?: string;
	} & HTMLAttributes<SVGSVGElement> = $props();
</script>

<svg
	xmlns="http://www.w3.org/2000/svg"
	width={size}
	height={size}
	viewBox="0 0 24 24"
	fill="none"
	stroke={color}
	stroke-width={strokeWidth}
	stroke-linecap="round"
	stroke-linejoin="round"
	class={cn('animate-spin', className)}
	role="status"
	aria-label="Loading"
	{...rest}
>
	<path d="M21 12a9 9 0 1 1-6.219-8.56" />
</svg>