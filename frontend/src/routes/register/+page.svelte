<script lang="ts">
	import { goto } from '$app/navigation';
	import { login, register } from '$lib/api/auth';
	import { setUser } from '$lib/stores/auth';
	import { ApiError } from '$lib/api/client';
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

<div class="mx-auto max-w-sm">
	<h1 class="text-2xl font-semibold mb-6">{m.register_heading()}</h1>
	<form onsubmit={submit} class="space-y-4">
		<label class="block">
			<span class="text-sm text-slate-700">{m.username()}</span>
			<input
				type="text"
				bind:value={username}
				required
				minlength="3"
				maxlength="50"
				class="mt-1 block w-full rounded border border-slate-300 px-3 py-2"
			/>
		</label>
		<label class="block">
			<span class="text-sm text-slate-700">{m.email()}</span>
			<input
				type="email"
				bind:value={email}
				required
				class="mt-1 block w-full rounded border border-slate-300 px-3 py-2"
			/>
		</label>
		<label class="block">
			<span class="text-sm text-slate-700">{m.password()}</span>
			<input
				type="password"
				bind:value={password}
				required
				minlength="8"
				class="mt-1 block w-full rounded border border-slate-300 px-3 py-2"
			/>
		</label>
		{#if error}
			<p class="text-sm text-red-600">{error}</p>
		{/if}
		<button
			type="submit"
			disabled={busy}
			class="w-full rounded bg-slate-900 px-4 py-2 text-white disabled:opacity-60"
		>
			{busy ? m.registering() : m.register_btn()}
		</button>
	</form>
	<p class="mt-4 text-sm text-slate-600">
		{m.has_account()} <a href="/login" class="text-slate-900 underline">{m.login_link()}</a>
	</p>
</div>
