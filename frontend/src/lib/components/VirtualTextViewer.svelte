<script lang="ts">
	import { onMount } from 'svelte';

	type TextRow = {
		key: string;
		text: string;
		lineNumber: number;
		continuation: boolean;
	};

	const ROW_HEIGHT = 22;
	const OVERSCAN_ROWS = 16;
	const MAX_ROW_CHARS = 4000;
	const FALLBACK_VIEWPORT_HEIGHT = 640;
	const EMPTY_LINE = String.fromCharCode(160);

	let { content, ariaLabel = 'Text preview', wrap = false }: { content: string; ariaLabel?: string; wrap?: boolean } = $props();

	let viewport = $state<HTMLDivElement | undefined>(undefined);
	let scrollTop = $state(0);
	let viewportHeight = $state(FALLBACK_VIEWPORT_HEIGHT);

	let rows = $derived(buildRows(content));
	let totalHeight = $derived(Math.max(rows.length * ROW_HEIGHT, ROW_HEIGHT));
	let startIndex = $derived(Math.max(0, Math.floor(scrollTop / ROW_HEIGHT) - OVERSCAN_ROWS));
	let visibleRowCount = $derived(Math.ceil(viewportHeight / ROW_HEIGHT) + OVERSCAN_ROWS * 2);
	let endIndex = $derived(Math.min(rows.length, startIndex + visibleRowCount));
	let offsetY = $derived(startIndex * ROW_HEIGHT);
	let visibleRows = $derived(rows.slice(startIndex, endIndex));

	function buildRows(value: string): TextRow[] {
		const sourceLines = value.split(/\r\n|\r|\n/);
		const nextRows: TextRow[] = [];

		for (const [lineIndex, line] of sourceLines.entries()) {
			const lineNumber = lineIndex + 1;

			if (line.length <= MAX_ROW_CHARS) {
				nextRows.push({
					key: `${lineIndex}:0`,
					text: line,
					lineNumber,
					continuation: false
				});
				continue;
			}

			for (let offset = 0; offset < line.length; offset += MAX_ROW_CHARS) {
				nextRows.push({
					key: `${lineIndex}:${offset}`,
					text: line.slice(offset, offset + MAX_ROW_CHARS),
					lineNumber,
					continuation: offset > 0
				});
			}
		}

		return nextRows;
	}

	function handleScroll() {
		scrollTop = viewport?.scrollTop ?? 0;
	}

	function updateViewportHeight() {
		viewportHeight = viewport?.clientHeight || FALLBACK_VIEWPORT_HEIGHT;
	}

	$effect(() => {
		content;
		wrap;
		if (!viewport) return;
		viewport.scrollTop = 0;
		scrollTop = 0;
		updateViewportHeight();
	});

	onMount(() => {
		updateViewportHeight();

		if (typeof ResizeObserver === 'undefined' || !viewport) {
			window.addEventListener('resize', updateViewportHeight);
			return () => window.removeEventListener('resize', updateViewportHeight);
		}

		const resizeObserver = new ResizeObserver(updateViewportHeight);
		resizeObserver.observe(viewport);

		return () => resizeObserver.disconnect();
	});
</script>

{#if wrap}
	<div role="region" aria-label={ariaLabel} class="bg-white font-mono text-sm text-ink-2">
		{#each rows as row (row.key)}
			<div class="flex items-start" style={`min-height: ${ROW_HEIGHT}px; line-height: ${ROW_HEIGHT}px;`}>
				<span class="sticky left-0 z-10 w-16 shrink-0 select-none border-r border-line-soft bg-surface-muted px-3 text-right text-xs text-ink-4">
					{row.continuation ? '·' : row.lineNumber}
				</span>
				<code class="block flex-1 whitespace-pre-wrap break-words px-4">{row.text || EMPTY_LINE}</code>
			</div>
		{/each}
	</div>
{:else}
	<div
		bind:this={viewport}
		role="region"
		aria-label={ariaLabel}
		onscroll={handleScroll}
		class="max-h-[80vh] overflow-auto bg-white font-mono text-sm text-ink-2"
	>
		<div class="relative min-w-full" style={`height: ${totalHeight}px;`}>
			<div class="absolute left-0 top-0 min-w-full" style={`transform: translateY(${offsetY}px);`}>
				{#each visibleRows as row (row.key)}
					<div class="flex min-w-max" style={`height: ${ROW_HEIGHT}px; line-height: ${ROW_HEIGHT}px;`}>
						<span class="sticky left-0 z-10 w-16 shrink-0 select-none border-r border-line-soft bg-surface-muted px-3 text-right text-xs text-ink-4">
							{row.continuation ? '·' : row.lineNumber}
						</span>
						<code class="block whitespace-pre px-4">{row.text || EMPTY_LINE}</code>
					</div>
				{/each}
			</div>
		</div>
	</div>
{/if}
