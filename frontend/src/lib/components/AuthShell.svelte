<script lang="ts">
	import type { Snippet } from 'svelte';
	import { HardDrive } from '@lucide/svelte';
	import LanguageDropdown from '$lib/components/LanguageDropdown.svelte';

	let { children }: { children: Snippet } = $props();
</script>

<svelte:head>
	<style>
		.auth-shell {
			position: relative;
			height: 100dvh;
			min-height: 600px;
			background:
				radial-gradient(ellipse 55% 50% at 38% 30%, color-mix(in srgb, var(--color-primary) 14%, transparent) 0%, transparent 55%),
				radial-gradient(ellipse 70% 55% at 35% 50%, color-mix(in srgb, var(--color-primary) 7%, transparent) 0%, transparent 55%),
				var(--color-surface-muted);
			color: var(--color-ink);
		}

		.auth-shell__inner {
			position: relative;
			z-index: 10;
			display: flex;
			height: 100%;
			flex-direction: column;
			padding: 16px 20px;
		}

		.auth-shell__bar,
		.auth-shell__content {
			margin-inline: auto;
			width: 100%;
			max-width: 64rem;
		}

		.auth-shell__bar {
			display: flex;
			align-items: center;
			justify-content: space-between;
			gap: 16px;
			flex-shrink: 0;
		}

		.auth-shell__content {
			display: flex;
			flex: 1 1 auto;
			min-height: 0;
		}

		.auth-grid {
			display: grid;
			width: 100%;
			flex: 1 1 auto;
			align-items: center;
			gap: 32px;
			padding-block: 16px;
			min-height: 0;
		}

		.auth-card-wrap {
			margin-inline: auto;
			width: 100%;
			max-width: 420px;
		}

		.auth-card {
			width: 100%;
			border-radius: 12px;
		}

		@media (min-width: 640px) {
			.auth-shell__inner {
				padding-inline: 32px;
			}
		}

		@media (min-width: 1024px) {
			.auth-shell__inner {
				padding: 20px 40px;
			}

			.auth-grid {
				gap: 48px;
				padding-block: 16px;
			}

			.auth-grid--login {
				grid-template-columns: minmax(0, 1fr) 400px;
			}

			.auth-grid--register {
				grid-template-columns: minmax(0, 0.9fr) 420px;
			}
		}
	</style>
</svelte:head>

<div class="auth-shell">
	<div class="bg-surface-muted"></div>
	<div class="absolute inset-0 opacity-[0.04] [background-image:linear-gradient(color-mix(in_srgb,var(--color-ink)_8%,transparent)_1px,transparent_1px),linear-gradient(90deg,color-mix(in_srgb,var(--color-ink)_8%,transparent)_1px,transparent_1px)] [background-size:42px_42px]"></div>
	<div class="auth-shell__inner">
		<div class="auth-shell__bar">
			<a href="/" class="flex items-center gap-2 text-sm font-semibold text-ink">
				<span class="flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-primary-on">
					<HardDrive size={18} />
				</span>
				<span>Netdisk</span>
			</a>
			<LanguageDropdown
				triggerClass="flex items-center gap-1.5 rounded-full border border-line bg-surface/80 px-3 py-2 text-xs font-medium text-ink-3 shadow-pop backdrop-blur transition-colors hover:bg-surface hover:text-ink data-[state=open]:bg-surface data-[state=open]:text-ink"
				contentClass="min-w-[124px]"
			/>
		</div>

		<div class="auth-shell__content">
			{@render children()}
		</div>
	</div>
</div>
