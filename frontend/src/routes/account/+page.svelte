<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { user, setUser, authReady } from '$lib/stores/auth';
	import { HardDrive, Camera, Save, User, Pencil, X, Shield, Calendar, Crown } from '@lucide/svelte';
	import { fmtSize } from '$lib/utils/format';
	import { getProfile, updateProfile, uploadAvatar, type ProfileData } from '$lib/api/profile';
	import * as m from '$lib/paraglide/messages';

	let profile = $state<ProfileData | null>(null);
	let loading = $state(true);
	let saving = $state(false);
	let saveMsg = $state('');
	let editing = $state(false);

	// Editable fields
	let nickname = $state('');
	let bio = $state('');
	let avatarPreview = $state<string | null>(null);
	let avatarFile: File | null = null;

	onMount(() => {
		if (!browser) return;
		if (!$user) {
			void goto('/login');
			return;
		}
		getProfile()
			.then((p) => {
				profile = p;
				nickname = p.nickname;
				bio = p.bio;
			})
			.catch(() => {})
			.finally(() => (loading = false));
	});

	function startEdit() {
		if (!profile) return;
		nickname = profile.nickname;
		bio = profile.bio;
		avatarPreview = null;
		avatarFile = null;
		saveMsg = '';
		editing = true;
	}

	function cancelEdit() {
		editing = false;
		avatarPreview = null;
		avatarFile = null;
		saveMsg = '';
	}

	function onAvatarSelect(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		avatarFile = file;
		const reader = new FileReader();
		reader.onload = () => (avatarPreview = reader.result as string);
		reader.readAsDataURL(file);
	}

	async function handleSave() {
		if (!profile) return;
		saving = true;
		saveMsg = '';
		try {
			if (avatarFile) {
				const res = await uploadAvatar(avatarFile);
				profile.avatar_url = res.avatar_url;
				avatarFile = null;
			}
			const updated = await updateProfile({ nickname, bio });
			profile = updated;
			if ($user) {
				setUser({
					...$user,
					...updated,
				});
			}
			saveMsg = m.profile_saved();
			editing = false;
		} catch {
			saveMsg = m.profile_save_failed();
		} finally {
			saving = false;
		}
	}

	let usedBytes = $derived($user?.storage?.storage_used ?? 0);
	let quotaBytes = $derived($user?.storage?.storage_quota ?? 0);

	function storagePercent(): number {
		if (quotaBytes <= 0) return 0;
		return Math.min((usedBytes / quotaBytes) * 100, 100);
	}

	function fmtDate(iso: string): string {
		try {
			return new Date(iso).toLocaleDateString();
		} catch {
			return iso;
		}
	}
</script>

{#if !$authReady}
	<!-- Wait for client-side auth check to avoid SSR flash -->
{:else if $user}
	<div class="space-y-6">
		<h1 class="text-xl font-semibold">{m.account_center()}</h1>

		<!-- Profile -->
		<div class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm">
			<div class="mb-4 flex items-center justify-between">
				<h2 class="text-sm font-medium text-gray-500">{m.profile_info()}</h2>
				{#if !editing}
					<button
						onclick={startEdit}
						class="inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm text-gray-600 transition-colors hover:bg-gray-50 hover:text-gray-900"
					>
						<Pencil size={14} />
						{m.edit()}
					</button>
				{/if}
			</div>

			{#if editing}
				<!-- Edit mode -->
				<div class="flex flex-col gap-6 sm:flex-row">
					<div class="flex flex-col items-center gap-3">
						<div class="relative">
							{#if avatarPreview || profile?.avatar_url}
								<img
									src={avatarPreview || profile?.avatar_url}
									alt="avatar"
									class="h-24 w-24 rounded-full object-cover ring-2 ring-gray-100"
								/>
							{:else}
								<div class="flex h-24 w-24 items-center justify-center rounded-full bg-gray-100 ring-2 ring-gray-100">
									<User size={40} class="text-gray-400" />
								</div>
							{/if}
							<label
								class="absolute -bottom-1 -right-1 flex h-8 w-8 cursor-pointer items-center justify-center rounded-full bg-gray-900 text-white shadow-sm hover:bg-gray-700 transition-colors"
							>
								<Camera size={14} />
								<input type="file" accept="image/jpeg,image/png,image/webp" class="hidden" onchange={onAvatarSelect} />
							</label>
						</div>
						<p class="text-xs text-gray-400">{m.avatar_hint()}</p>
					</div>

					<div class="flex-1 space-y-4">
						<div>
							<label class="mb-1 block text-sm font-medium text-gray-700" for="nickname">{m.nickname_label()}</label>
							<input
								id="nickname"
								type="text"
								bind:value={nickname}
								placeholder={$user.username}
								maxlength={100}
								class="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none"
							/>
						</div>
						<div>
							<label class="mb-1 block text-sm font-medium text-gray-700" for="bio">{m.bio_label()}</label>
							<textarea
								id="bio"
								bind:value={bio}
								placeholder={m.bio_placeholder()}
								rows={3}
								maxlength={500}
								class="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none resize-none"
							></textarea>
						</div>

						<div class="flex items-center gap-3">
							<button
								onclick={handleSave}
								disabled={saving}
								class="inline-flex items-center gap-2 rounded-lg bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-700 disabled:opacity-50 transition-colors"
							>
								<Save size={14} />
								{saving ? m.saving() : m.save()}
							</button>
							<button
								onclick={cancelEdit}
								class="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 px-4 py-2 text-sm text-gray-600 transition-colors hover:bg-gray-50"
							>
								<X size={14} />
								{m.cancel()}
							</button>
							{#if saveMsg}
								<span class="text-sm {saveMsg === m.profile_saved() ? 'text-green-600' : 'text-red-600'}">{saveMsg}</span>
							{/if}
						</div>
					</div>
				</div>
			{:else}
				<!-- View mode -->
				<div class="flex flex-col gap-6 sm:flex-row">
					<div class="flex flex-col items-center gap-3">
						{#if profile?.avatar_url}
							<img
								src={profile.avatar_url}
								alt="avatar"
								class="h-24 w-24 rounded-full object-cover ring-2 ring-gray-100"
							/>
						{:else}
							<div class="flex h-24 w-24 items-center justify-center rounded-full bg-gray-100 ring-2 ring-gray-100">
								<User size={40} class="text-gray-400" />
							</div>
						{/if}
					</div>

					<div class="flex-1 space-y-3">
						<div>
							<p class="text-xs font-medium text-gray-400 mb-1">{m.nickname_label()}</p>
							<p class="text-sm text-gray-800">{profile?.nickname || $user.username}</p>
						</div>
						<div>
							<p class="text-xs font-medium text-gray-400 mb-1">{m.bio_label()}</p>
							<p class="text-sm text-gray-800 whitespace-pre-wrap">{profile?.bio || m.no_bio()}</p>
						</div>
					</div>
				</div>
			{/if}
		</div>

		<!-- Account info -->
		<div class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm">
			<h2 class="mb-4 text-sm font-medium text-gray-500">{m.account_info()}</h2>
			<div class="grid gap-4 sm:grid-cols-2">
				<div class="flex items-center gap-3">
					<div class="flex h-9 w-9 items-center justify-center rounded-lg bg-gray-50">
						<User size={16} class="text-gray-400" />
					</div>
					<div>
						<p class="text-xs text-gray-400">{m.username_label()}</p>
						<p class="text-sm font-medium text-gray-800">{$user.username}</p>
					</div>
				</div>
				<div class="flex items-center gap-3">
					<div class="flex h-9 w-9 items-center justify-center rounded-lg bg-gray-50">
						<Shield size={16} class="text-gray-400" />
					</div>
					<div>
						<p class="text-xs text-gray-400">{m.email_label()}</p>
						<p class="text-sm font-medium text-gray-800">{$user.email}</p>
					</div>
				</div>
				<div class="flex items-center gap-3">
					<div class="flex h-9 w-9 items-center justify-center rounded-lg bg-gray-50">
						<Crown size={16} class="text-gray-400" />
					</div>
					<div>
						<p class="text-xs text-gray-400">{m.level()}</p>
						<p class="text-sm font-medium text-gray-800">{$user.level?.level_name || '-'}</p>
						{#if $user.level?.expires_at}
							<p class="text-xs text-gray-400">{m.level_expires({ date: fmtDate($user.level.expires_at) })}</p>
						{/if}
					</div>
				</div>
				<div class="flex items-center gap-3">
					<div class="flex h-9 w-9 items-center justify-center rounded-lg bg-gray-50">
						<Calendar size={16} class="text-gray-400" />
					</div>
					<div>
						<p class="text-xs text-gray-400">{m.joined()}</p>
						<p class="text-sm font-medium text-gray-800">{fmtDate($user.created_at)}</p>
					</div>
				</div>
			</div>
		</div>

		<!-- Storage -->
		<div class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm">
			<h2 class="mb-4 flex items-center gap-2 text-sm font-medium text-gray-500">
				<HardDrive size={16} /> {m.drive_storage()}
			</h2>
			<div class="space-y-3">
				{#if loading}
					<p class="text-sm text-gray-400">{m.loading()}</p>
				{:else}
					<!-- Usage bar -->
					<div>
						<div class="mb-2 flex items-baseline justify-between">
							<span class="text-2xl font-semibold text-gray-900">{fmtSize(usedBytes)}</span>
							<span class="text-sm text-gray-400">/ {fmtSize(quotaBytes)}</span>
						</div>
						<div class="h-2.5 w-full overflow-hidden rounded-full bg-gray-100">
							<div
								class="h-full rounded-full transition-all {storagePercent() > 90 ? 'bg-red-500' : storagePercent() > 70 ? 'bg-amber-500' : 'bg-blue-600'}"
								style="width:{storagePercent()}%"
							></div>
						</div>
						<p class="mt-1.5 text-right text-xs text-gray-400">{storagePercent().toFixed(1)}%</p>
					</div>

					<p class="text-xs text-gray-400">{m.used()}: {fmtSize(usedBytes)} &middot; {m.drive_storage()}: {fmtSize(quotaBytes)}</p>
				{/if}
			</div>
		</div>
	</div>
{:else}
	<p class="text-gray-600">{@html m.please_login({ link: '<a href="/login" class="underline">' + m.login_link_text() + '</a>' })}</p>
{/if}
