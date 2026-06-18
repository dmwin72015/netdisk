<script lang="ts">
	import { goto } from '$app/navigation';
	import { login, register } from '$lib/api/auth';
	import { setUser } from '$lib/stores/auth';
	import { ApiError } from '$lib/api/client';
	import AuthShell from '$lib/components/AuthShell.svelte';
	import { AtSign, CircleCheck, KeyRound, Mail, ShieldCheck, Sparkles, UserRound } from '@lucide/svelte';
	import * as m from '$lib/paraglide/messages';

	let username = $state('');
	let email = $state('');
	let password = $state('');
	let error = $state<string | null>(null);
	let busy = $state(false);

	async function submit(e: Event) {
		e.preventDefault();
		error = null;
		busy = true;
		try {
			await register(username, email, password);
			const res = await login(email, password);
			setUser(res.user);
			await goto('/');
		} catch (err) {
			error = err instanceof ApiError ? err.message : m.register_failed();
		} finally {
			busy = false;
		}
	}
</script>

<AuthShell>
	<section class="auth-grid auth-grid--register grid w-full flex-1 items-center gap-8 py-4 lg:grid-cols-[minmax(0,0.9fr)_420px] lg:gap-12 lg:py-4">
		<div class="hidden max-w-xl lg:block">
			<div class="mb-5 inline-flex items-center gap-2 rounded-full border border-white/70 bg-white/70 px-3 py-1.5 text-xs font-medium text-slate-600 shadow-sm shadow-slate-200/80 backdrop-blur">
				<Sparkles size={14} class="text-blue-600" />
				<span>{m.register_kicker()}</span>
			</div>
			<h1 class="max-w-xl text-4xl font-semibold leading-tight tracking-tight text-slate-950 xl:text-[2.75rem]">
				{m.register_title()}
			</h1>
			<p class="mt-4 max-w-lg text-sm leading-6 text-slate-600">
				{m.register_subtitle()}
			</p>
			<div class="mt-6 rounded-lg border border-white/70 bg-white/75 p-3.5 shadow-sm shadow-slate-200/70 backdrop-blur">
				<div class="flex items-center justify-between border-b border-slate-200/70 pb-3">
					<div>
						<p class="text-sm font-semibold text-slate-950">{m.register_panel_title()}</p>
						<p class="mt-0.5 text-xs text-slate-500">{m.register_panel_subtitle()}</p>
					</div>
					<span class="rounded-full bg-blue-50 px-3 py-1 text-xs font-medium text-blue-600">{m.register_panel_badge()}</span>
				</div>
				<div class="mt-3 space-y-2.5">
					<div class="flex items-start gap-3 rounded-lg bg-white/80 p-2.5">
						<span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-blue-50 text-blue-600">
							<AtSign size={16} />
						</span>
						<div class="min-w-0 flex-1">
							<p class="text-sm font-semibold text-slate-900">{m.register_feature_account_title()}</p>
							<p class="mt-0.5 text-xs leading-5 text-slate-500">{m.register_feature_account_desc()}</p>
						</div>
						<span class="text-xs font-semibold text-slate-300">01</span>
					</div>
					<div class="flex items-start gap-3 rounded-lg bg-white/80 p-2.5">
						<span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-emerald-50 text-emerald-600">
							<ShieldCheck size={16} />
						</span>
						<div class="min-w-0 flex-1">
							<p class="text-sm font-semibold text-slate-900">{m.register_feature_security_title()}</p>
							<p class="mt-0.5 text-xs leading-5 text-slate-500">{m.register_feature_security_desc()}</p>
						</div>
						<span class="text-xs font-semibold text-slate-300">02</span>
					</div>
					<div class="flex items-start gap-3 rounded-lg bg-white/80 p-2.5">
						<span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-violet-50 text-violet-600">
							<CircleCheck size={16} />
						</span>
						<div class="min-w-0 flex-1">
							<p class="text-sm font-semibold text-slate-900">{m.register_feature_ready_title()}</p>
							<p class="mt-0.5 text-xs leading-5 text-slate-500">{m.register_feature_ready_desc()}</p>
						</div>
						<span class="text-xs font-semibold text-slate-300">03</span>
					</div>
				</div>
			</div>
		</div>

		<div class="auth-card-wrap mx-auto w-full max-w-[420px]">
			<div class="auth-card rounded-xl border border-white/80 bg-white/95 p-5 shadow-2xl shadow-slate-300/50 backdrop-blur-xl sm:p-6">
				<div class="mb-5">
					<div class="mb-3 flex h-10 w-10 items-center justify-center rounded-lg bg-blue-600 text-white shadow-lg shadow-blue-200">
						<UserRound size={20} />
					</div>
					<h1 class="text-2xl font-semibold tracking-tight text-slate-950">{m.register_heading()}</h1>
					<p class="mt-1.5 text-sm leading-6 text-slate-500">
						{m.has_account()}
						<a href="/login" class="font-medium text-blue-600 transition-colors hover:text-blue-700">
							{m.login_link()}
						</a>
					</p>
				</div>

				<form onsubmit={submit} class="space-y-3.5">
					<label class="block">
						<span class="text-sm font-medium text-slate-700">{m.username()}</span>
						<span class="mt-1.5 flex items-center gap-3 rounded-lg border border-slate-200 bg-slate-50/80 px-3.5 py-2.5 transition focus-within:border-blue-400 focus-within:bg-white focus-within:ring-4 focus-within:ring-blue-100">
							<UserRound size={16} class="shrink-0 text-slate-400" />
							<input
								type="text"
								bind:value={username}
								required
								minlength="3"
								maxlength="50"
								autocomplete="username"
								class="w-full bg-transparent text-[15px] text-slate-950 outline-none placeholder:text-slate-400"
							/>
						</span>
					</label>
					<label class="block">
						<span class="text-sm font-medium text-slate-700">{m.email()}</span>
						<span class="mt-1.5 flex items-center gap-3 rounded-lg border border-slate-200 bg-slate-50/80 px-3.5 py-2.5 transition focus-within:border-blue-400 focus-within:bg-white focus-within:ring-4 focus-within:ring-blue-100">
							<Mail size={16} class="shrink-0 text-slate-400" />
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
						<span class="mt-1.5 flex items-center gap-3 rounded-lg border border-slate-200 bg-slate-50/80 px-3.5 py-2.5 transition focus-within:border-blue-400 focus-within:bg-white focus-within:ring-4 focus-within:ring-blue-100">
							<KeyRound size={16} class="shrink-0 text-slate-400" />
							<input
								type="password"
								bind:value={password}
								required
								minlength="8"
								autocomplete="new-password"
								class="w-full bg-transparent text-[15px] text-slate-950 outline-none placeholder:text-slate-400"
							/>
						</span>
					</label>
					{#if error}
						<p class="rounded-lg border border-red-100 bg-red-50 px-3 py-2 text-sm text-red-600">{error}</p>
					{/if}
					<button
						type="submit"
						disabled={busy}
						class="flex w-full items-center justify-center rounded-lg bg-slate-950 px-4 py-2.5 text-sm font-semibold text-white shadow-lg shadow-slate-300 transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
					>
						{busy ? m.registering() : m.register_btn()}
					</button>
				</form>
			</div>
		</div>
	</section>
</AuthShell>
