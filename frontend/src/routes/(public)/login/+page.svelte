<script lang="ts">
	import { goto } from '$app/navigation';
	import { login } from '$lib/api/auth';
	import { setUser } from '$lib/stores/auth';
	import { ApiError, api, updateTokens, type Tokens, type UserInfo } from '$lib/api/client';
	import AuthShell from '$lib/components/AuthShell.svelte';
	import AlertDialog from '$lib/ui/alert-dialog/AlertDialog.svelte';
	import { Cloud, LockKeyhole, Mail, ShieldCheck, Sparkles, ExternalLink } from '@lucide/svelte';
	import * as m from '$lib/paraglide/messages';

	let email = $state('');
	let password = $state('');
	let error = $state<string | null>(null);
	let busy = $state(false);
	let oauthBusy = $state(false);
	let emailConfirmOpen = $state(false);
	let emailConfirmTarget = $state<{ token: string; email: string; provider: string } | null>(null);
	let emailConfirmBusy = $state(false);

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

	function oauthLogin(provider: string) {
		const width = 520;
		const height = 600;
		const left = (screen.width - width) / 2;
		const top = (screen.height - height) / 2;
		const popup = window.open(
			`/api/v1/auth/oauth/${provider}/authorize`,
			'oauth-popup',
			`width=${width},height=${height},left=${left},top=${top},popup=1`,
		);

		if (!popup) {
			error = m.oauth_popup_blocked();
			return;
		}

		oauthBusy = true;

		async function onMessage(event: MessageEvent) {
			if (event.origin !== location.origin) return;
			const data = event.data ?? {};
			const { accessToken, refreshToken, emailMatchConfirm, confirmToken, email: matchEmail, provider } = data;

			if (emailMatchConfirm && confirmToken) {
				popup?.close();
				emailConfirmTarget = { token: confirmToken, email: matchEmail, provider };
				emailConfirmOpen = true;
				return;
			}

			if (!accessToken || !refreshToken) return;

			window.removeEventListener('message', onMessage);
			updateTokens({ accessToken, refreshToken, expiresIn: 3600 });

			api<UserInfo>('/api/v1/user/me', { method: 'GET' })
				.then((data) => {
					setUser(data);
					goto('/');
				})
				.catch(() => {
					error = 'Failed to complete login';
				})
				.finally(() => {
					oauthBusy = false;
				});
		}

		window.addEventListener('message', onMessage);

		const closeCheck = setInterval(() => {
			if (popup.closed) {
				clearInterval(closeCheck);
				window.removeEventListener('message', onMessage);
				oauthBusy = false;
			}
		}, 500);
	}

	async function confirmEmailLink() {
		if (!emailConfirmTarget) return;
		emailConfirmBusy = true;
		error = null;
		try {
			const tokens = await api<Tokens>(
				`/api/v1/auth/oauth/email-confirm?token=${encodeURIComponent(emailConfirmTarget.token)}`,
				{ method: 'POST' },
			);
			emailConfirmTarget = null;
			updateTokens(tokens);
			const userData = await api<UserInfo>('/api/v1/user/me', { method: 'GET' });
			setUser(userData);
			await goto('/');
		} catch (err) {
			error = err instanceof ApiError ? err.message : m.oauth_link_failed();
		} finally {
			emailConfirmBusy = false;
		}
	}
</script>

<AuthShell>
	<section class="auth-grid auth-grid--login grid w-full flex-1 items-center gap-8 py-4 lg:grid-cols-[minmax(0,1fr)_400px] lg:gap-12 lg:py-4">
		<div class="hidden max-w-2xl lg:block">
			<div class="mb-5 inline-flex items-center gap-2 rounded-full border border-line bg-surface-sunken px-3 py-1.5 text-xs font-medium text-ink-3">
				<Sparkles size={14} class="text-primary" />
				<span>{m.login_kicker()}</span>
			</div>
			<h1 class="max-w-xl text-4xl font-semibold leading-tight tracking-tight text-ink xl:text-[2.75rem]">
				{m.login_title()}
			</h1>
			<p class="mt-4 max-w-lg text-sm leading-6 text-ink-3">
				{m.login_subtitle()}
			</p>
			<div class="mt-7 grid max-w-xl grid-cols-3 gap-3">
				<div class="rounded-lg border border-line bg-surface p-3.5">
					<Cloud size={18} class="text-primary" />
					<p class="mt-3 text-sm font-semibold text-ink">{m.login_feature_files_title()}</p>
					<p class="mt-1 text-xs leading-5 text-ink-3">{m.login_feature_files_desc()}</p>
				</div>
				<div class="rounded-lg border border-line bg-surface p-3.5">
					<ShieldCheck size={18} class="text-success" />
					<p class="mt-3 text-sm font-semibold text-ink">{m.login_feature_secure_title()}</p>
					<p class="mt-1 text-xs leading-5 text-ink-3">{m.login_feature_secure_desc()}</p>
				</div>
				<div class="rounded-lg border border-line bg-surface p-3.5">
					<LockKeyhole size={18} class="text-primary" />
					<p class="mt-3 text-sm font-semibold text-ink">{m.login_feature_fast_title()}</p>
					<p class="mt-1 text-xs leading-5 text-ink-3">{m.login_feature_fast_desc()}</p>
				</div>
			</div>
		</div>

		<div class="auth-card-wrap mx-auto w-full max-w-[420px]">
			<div class="auth-card rounded-xl border border-line bg-white p-5 shadow-dialog sm:p-6">
				<div class="mb-5">
					<div class="mb-3 flex h-10 w-10 items-center justify-center rounded-lg bg-primary text-primary-on">
						<LockKeyhole size={20} />
					</div>
					<h1 class="text-2xl font-semibold tracking-tight text-ink">{m.login_heading()}</h1>
					<p class="mt-1.5 text-sm leading-6 text-ink-3">
						{m.no_account()}
						<a href="/register" class="font-medium text-primary transition-colors hover:text-primary">
							{m.register_link()}
						</a>
					</p>
				</div>

				<form onsubmit={submit} class="space-y-3.5">
					<label class="block">
						<span class="text-sm font-medium text-ink-2">{m.email()}</span>
						<span class="mt-1.5 flex items-center gap-3 rounded-lg border border-line bg-surface-muted/80 px-3.5 py-2.5 transition focus-within:border-primary">
							<Mail size={16} class="shrink-0 text-ink-4" />
							<input
								type="email"
								bind:value={email}
								required
								autocomplete="email"
								class="w-full bg-transparent text-body text-ink outline-none placeholder:text-ink-4"
							/>
						</span>
					</label>
					<label class="block">
						<span class="text-sm font-medium text-ink-2">{m.password()}</span>
						<span class="mt-1.5 flex items-center gap-3 rounded-lg border border-line bg-surface-muted/80 px-3.5 py-2.5 transition focus-within:border-primary">
							<LockKeyhole size={16} class="shrink-0 text-ink-4" />
							<input
								type="password"
								bind:value={password}
								required
								minlength="8"
								autocomplete="current-password"
								class="w-full bg-transparent text-body text-ink outline-none placeholder:text-ink-4"
							/>
						</span>
					</label>
					{#if error}
						<p class="rounded-lg border border-danger bg-danger-soft px-3 py-2 text-sm text-danger">{error}</p>
					{/if}
					<button
						type="submit"
						disabled={busy}
						class="flex w-full items-center justify-center rounded-lg bg-ink px-4 py-2.5 text-sm font-semibold text-primary-on transition hover:bg-ink-2 disabled:cursor-not-allowed disabled:opacity-60"
					>
						{busy ? m.logging_in() : m.login_btn()}
					</button>
				</form>

				<div class="relative mt-4">
					<div class="absolute inset-0 flex items-center">
						<div class="w-full border-t border-line"></div>
					</div>
					<div class="relative flex justify-center text-xs">
						<span class="bg-white px-2 text-ink-4">or continue with</span>
					</div>
				</div>

				<div class="mt-4 grid grid-cols-2 gap-2.5">
					<button
						onclick={() => oauthLogin('2libra')}
						disabled={oauthBusy}
						class="flex w-full items-center justify-center gap-2 rounded-lg border border-line bg-white px-3 py-2.5 text-sm font-medium text-ink-2 transition hover:bg-surface-muted disabled:cursor-not-allowed disabled:opacity-60"
					>
						{#if oauthBusy}
							<ExternalLink size={16} class="animate-pulse" />
						{:else}
							<img src="/2libra.png" alt="" loading="lazy" class="h-4 w-4" />
						{/if}
						2libra
					</button>
					<button
						onclick={() => oauthLogin('github')}
						disabled={oauthBusy}
						class="flex w-full items-center justify-center gap-2 rounded-lg border border-line bg-white px-3 py-2.5 text-sm font-medium text-ink-2 transition hover:bg-surface-muted disabled:cursor-not-allowed disabled:opacity-60"
					>
						{#if oauthBusy}
							<ExternalLink size={16} class="animate-pulse" />
						{:else}
							<svg viewBox="0 0 24 24" class="h-4 w-4" fill="currentColor"><path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"/></svg>
						{/if}
						GitHub
					</button>
				</div>
			</div>
		</div>
	</section>
</AuthShell>

<AlertDialog
	bind:open={emailConfirmOpen}
	title={m.oauth_link_existing_title()}
	description={emailConfirmTarget
		? m.oauth_link_existing_desc({ email: emailConfirmTarget.email, provider: emailConfirmTarget.provider })
		: ''}
	confirmText={emailConfirmBusy ? m.oauth_linking() : m.oauth_link_login()}
	onConfirm={confirmEmailLink}
	onCancel={() => { emailConfirmOpen = false; emailConfirmTarget = null; }}
/>
