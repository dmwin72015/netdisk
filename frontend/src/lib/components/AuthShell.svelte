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
			min-height: 100vh;
			overflow: hidden;
			background: #f5f7fb;
			color: #020617;
		}

		.auth-shell__inner {
			position: relative;
			z-index: 10;
			display: flex;
			min-height: 100vh;
			flex-direction: column;
			padding: 20px;
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
		}

		.auth-shell__content {
			display: flex;
			flex: 1 1 auto;
		}

		.auth-grid {
			display: grid;
			width: 100%;
			flex: 1 1 auto;
			align-items: center;
			gap: 40px;
			padding-block: 32px;
		}

		.auth-card-wrap {
			margin-inline: auto;
			width: 100%;
			max-width: 420px;
		}

		.auth-card {
			width: 100%;
			border-radius: 8px;
		}

		@media (min-width: 640px) {
			.auth-shell__inner {
				padding-inline: 32px;
			}
		}

		@media (min-width: 1024px) {
			.auth-shell__inner {
				padding-inline: 40px;
			}

			.auth-grid {
				gap: 48px;
				padding-block: 48px;
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

<div class="auth-shell relative min-h-screen overflow-hidden bg-[#f5f7fb] text-slate-950">
	<div class="absolute inset-0 bg-[radial-gradient(circle_at_15%_10%,rgba(59,130,246,0.16),transparent_28%),radial-gradient(circle_at_85%_20%,rgba(20,184,166,0.12),transparent_26%),linear-gradient(135deg,#f8fafc_0%,#eef4ff_46%,#f6fbf9_100%)]"></div>
	<div class="absolute left-0 top-0 h-full w-full opacity-[0.07] [background-image:linear-gradient(#0f172a_1px,transparent_1px),linear-gradient(90deg,#0f172a_1px,transparent_1px)] [background-size:42px_42px]"></div>

	<div class="auth-shell__inner relative z-10 flex min-h-screen flex-col px-5 py-5 sm:px-8 lg:px-10">
		<div class="auth-shell__bar mx-auto flex w-full max-w-5xl items-center justify-between gap-4">
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

		<div class="auth-shell__content mx-auto flex w-full max-w-5xl flex-1">
			{@render children()}
		</div>
	</div>
</div>
