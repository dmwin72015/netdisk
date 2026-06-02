<script lang="ts">
	import { goto } from '$app/navigation';
	import { login } from '$lib/api/auth';
	import { setUser } from '$lib/stores/auth';
	import { ApiError } from '$lib/api/client';
	import AuthShell from '$lib/components/AuthShell.svelte';
	import { Cloud, LockKeyhole, Mail, ShieldCheck, Sparkles } from '@lucide/svelte';
	import * as m from '$lib/paraglide/messages';

	let email = $state('');
	let password = $state('');
	let error = $state<string | null>(null);
	let busy = $state(false);

	async function submit(e: Event) {
		e.preventDefault();
		error = null;
		busy = true;
		try {
			const res = await login(email, password);
			setUser(res.user);
			await goto('/');
		} catch (err) {
			error = err instanceof ApiError ? err.message : m.login_failed();
		} finally {
			busy = false;
		}
	}
</script>

<AuthShell>
	<section class="auth-grid auth-grid--login grid w-full flex-1 items-center gap-10 py-8 lg:grid-cols-[minmax(0,1fr)_400px] lg:gap-12 lg:py-12">
		<div class="hidden max-w-2xl lg:block">
			<div class="mb-7 inline-flex items-center gap-2 rounded-full border border-white/70 bg-white/70 px-3 py-1.5 text-xs font-medium text-slate-600 shadow-sm shadow-slate-200/80 backdrop-blur">
				<Sparkles size={14} class="text-blue-600" />
				<span>{m.login_kicker()}</span>
			</div>
			<h1 class="max-w-xl text-5xl font-semibold leading-tight text-slate-950">
				{m.login_title()}
			</h1>
			<p class="mt-5 max-w-lg text-base leading-7 text-slate-600">
				{m.login_subtitle()}
			</p>
			<div class="mt-10 grid max-w-xl grid-cols-3 gap-3">
				<div class="rounded-lg border border-white/70 bg-white/70 p-4 shadow-sm shadow-slate-200/70 backdrop-blur">
					<Cloud size={20} class="text-blue-600" />
					<p class="mt-4 text-sm font-semibold text-slate-900">{m.login_feature_files_title()}</p>
					<p class="mt-1 text-xs leading-5 text-slate-500">{m.login_feature_files_desc()}</p>
				</div>
				<div class="rounded-lg border border-white/70 bg-white/70 p-4 shadow-sm shadow-slate-200/70 backdrop-blur">
					<ShieldCheck size={20} class="text-emerald-600" />
					<p class="mt-4 text-sm font-semibold text-slate-900">{m.login_feature_secure_title()}</p>
					<p class="mt-1 text-xs leading-5 text-slate-500">{m.login_feature_secure_desc()}</p>
				</div>
				<div class="rounded-lg border border-white/70 bg-white/70 p-4 shadow-sm shadow-slate-200/70 backdrop-blur">
					<LockKeyhole size={20} class="text-violet-600" />
					<p class="mt-4 text-sm font-semibold text-slate-900">{m.login_feature_fast_title()}</p>
					<p class="mt-1 text-xs leading-5 text-slate-500">{m.login_feature_fast_desc()}</p>
				</div>
			</div>
		</div>

		<div class="auth-card-wrap mx-auto w-full max-w-[420px]">
			<div class="auth-card rounded-lg border border-white/80 bg-white/90 p-6 shadow-2xl shadow-slate-300/50 backdrop-blur-xl sm:p-8">
				<div class="mb-8">
					<div class="mb-5 flex h-12 w-12 items-center justify-center rounded-lg bg-blue-600 text-white shadow-lg shadow-blue-200">
						<LockKeyhole size={22} />
					</div>
					<h1 class="text-3xl font-semibold tracking-normal text-slate-950">{m.login_heading()}</h1>
					<p class="mt-2 text-sm leading-6 text-slate-500">
						{m.no_account()}
						<a href="/register" class="font-medium text-blue-600 transition-colors hover:text-blue-700">
							{m.register_link()}
						</a>
					</p>
				</div>

				<form onsubmit={submit} class="space-y-5">
					<label class="block">
						<span class="text-sm font-medium text-slate-700">{m.email()}</span>
						<span class="mt-2 flex items-center gap-3 rounded-lg border border-slate-200 bg-slate-50/80 px-4 py-3 transition focus-within:border-blue-400 focus-within:bg-white focus-within:ring-4 focus-within:ring-blue-100">
							<Mail size={18} class="shrink-0 text-slate-400" />
							<input
								type="email"
								bind:value={email}
								required
								autocomplete="email"
								class="w-full bg-transparent text-[15px] text-slate-950 outline-none placeholder:text-slate-400"
							/>
						</span>
					</label>
					<label class="block">
						<span class="text-sm font-medium text-slate-700">{m.password()}</span>
						<span class="mt-2 flex items-center gap-3 rounded-lg border border-slate-200 bg-slate-50/80 px-4 py-3 transition focus-within:border-blue-400 focus-within:bg-white focus-within:ring-4 focus-within:ring-blue-100">
							<LockKeyhole size={18} class="shrink-0 text-slate-400" />
							<input
								type="password"
								bind:value={password}
								required
								minlength="8"
								autocomplete="current-password"
								class="w-full bg-transparent text-[15px] text-slate-950 outline-none placeholder:text-slate-400"
							/>
						</span>
					</label>
					{#if error}
						<p class="rounded-lg border border-red-100 bg-red-50 px-4 py-3 text-sm text-red-600">{error}</p>
					{/if}
					<button
						type="submit"
						disabled={busy}
						class="flex w-full items-center justify-center rounded-lg bg-slate-950 px-4 py-3.5 text-sm font-semibold text-white shadow-lg shadow-slate-300 transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
					>
						{busy ? m.logging_in() : m.login_btn()}
					</button>
				</form>
			</div>
		</div>
	</section>
</AuthShell>
