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
			overflow: hidden;
			background: #f5f7fb;
			color: #020617;
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
			overflow: hidden;
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
	<div class="absolute inset-0 bg-[radial-gradient(circle_at_15%_10%,rgba(59,130,246,0.16),transparent_28%),radial-gradient(circle_at_85%_20%,rgba(20,184,166,0.12),transparent_26%),linear-gradient(135deg,#f8fafc_0%,#eef4ff_46%,#f6fbf9_100%)]"></div>
	<div class="absolute left-0 top-0 h-full w-full opacity-[0.07] [background-image:linear-gradient(#0f172a_1px,transparent_1px),linear-gradient(90deg,#0f172a_1px,transparent_1px)] [background-size:42px_42px]"></div>

	<div class="auth-shell__inner">
		<div class="auth-shell__bar">
			<a href="/" class="flex items-center gap-2 text-sm font-semibold text-slate-900">
				<span class="flex h-9 w-9 items-center justify-center rounded-lg bg-slate-950 text-white shadow-lg shadow-slate-300/70">
					<HardDrive size={18} />
				</span>
				<span>Netdisk</span>
			</a>
			<LanguageDropdown
				triggerClass="flex items-center gap-1.5 rounded-full border border-white/70 bg-white/80 px-3 py-2 text-xs font-medium text-slate-600 shadow-sm shadow-slate-200/70 backdrop-blur transition-colors hover:bg-white hover:text-slate-950 data-[state=open]:bg-white data-[state=open]:text-slate-950"
				contentClass="min-w-[124px]"
			/>
		</div>

		<div class="auth-shell__content">
			{@render children()}
		</div>
	</div>
</div>
