<script lang="ts">
	import { browser } from '$app/environment';

	let {
		src,
		alt = '',
		containerClass = '',
		imgClass = '',
		onError
	}: {
		src: string;
		alt?: string;
		containerClass?: string;
		imgClass?: string;
		onError?: () => void;
	} = $props();

	let container: HTMLSpanElement | undefined = $state();
	let shouldLoad = $state(false);
	let currentSrc = $state('');

	$effect(() => {
		if (src === currentSrc) return;
		currentSrc = src;
		shouldLoad = false;
	});

	$effect(() => {
		const node = container;
		if (!node || shouldLoad) return;

		if (!browser || !('IntersectionObserver' in window)) {
			shouldLoad = true;
			return;
		}

		const observer = new IntersectionObserver(
			(entries) => {
				if (!entries.some((entry) => entry.isIntersecting)) return;
				shouldLoad = true;
				observer.disconnect();
			},
			{ rootMargin: '240px 0px' }
		);

		observer.observe(node);
		return () => observer.disconnect();
	});
</script>

<span bind:this={container} class={containerClass}>
	{#if shouldLoad}
		<img
			src={currentSrc}
			{alt}
			loading="lazy"
			decoding="async"
			fetchpriority="low"
			class={imgClass}
			onerror={() => onError?.()}
		/>
	{/if}
</span>
