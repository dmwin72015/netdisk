<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { user, setUser, authReady } from '$lib/stores/auth';
	import { HardDrive, Camera, Save, User, Pencil, X } from '@lucide/svelte';
	import { driveStats, fmtSize } from '$lib/api/drive';
	import { getProfile, updateProfile, uploadAvatar, type ProfileData } from '$lib/api/profile';
	import * as m from '$lib/paraglide/messages';

	let stats = $state<{ used_bytes: number; base_bytes: number; member_bonus_bytes: number; pack_bytes: number; total_bytes: number } | null>(null);
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
		Promise.all([
			driveStats().then((r) => (stats = r)),
			getProfile().then((p) => {
				profile = p;
				nickname = p.nickname;
				bio = p.bio;
			})
		])
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
</script>

{#if !$authReady}
	<!-- Wait for client-side auth check to avoid SSR flash -->
{:else if $user}
	<div class="space-y-6">
		<h1 class="text-xl font-semibold">{m.account_center()}</h1>

		<!-- Profile -->
		<div class="rounded-lg border bg-white p-6">
			<div class="mb-4 flex items-center justify-between">
				<h2 class="text-sm font-medium text-slate-500">{m.profile_info()}</h2>
				{#if !editing}
					<button
						onclick={startEdit}
						class="inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm text-slate-600 transition-colors hover:bg-slate-50 hover:text-slate-900"
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
									class="h-24 w-24 rounded-full object-cover ring-2 ring-slate-100"
								/>
							{:else}
								<div class="flex h-24 w-24 items-center justify-center rounded-full bg-slate-100 ring-2 ring-slate-100">
									<User size={40} class="text-slate-400" />
								</div>
							{/if}
							<label
								class="absolute -bottom-1 -right-1 flex h-8 w-8 cursor-pointer items-center justify-center rounded-full bg-slate-900 text-white shadow-sm hover:bg-slate-700 transition-colors"
							>
								<Camera size={14} />
								<input type="file" accept="image/jpeg,image/png,image/webp" class="hidden" onchange={onAvatarSelect} />
							</label>
						</div>
						<p class="text-xs text-slate-400">{m.avatar_hint()}</p>
					</div>

					<div class="flex-1 space-y-4">
						<div>
							<label class="mb-1 block text-sm font-medium text-slate-700" for="nickname">{m.nickname_label()}</label>
							<input
								id="nickname"
								type="text"
								bind:value={nickname}
								placeholder={$user.username}
								maxlength={100}
								class="w-full rounded-lg border border-slate-200 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none"
							/>
						</div>
						<div>
							<label class="mb-1 block text-sm font-medium text-slate-700" for="bio">{m.bio_label()}</label>
							<textarea
								id="bio"
								bind:value={bio}
								placeholder={m.bio_placeholder()}
								rows={3}
								maxlength={500}
								class="w-full rounded-lg border border-slate-200 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none resize-none"
							></textarea>
						</div>

						<div class="flex items-center gap-3">
							<button
								onclick={handleSave}
								disabled={saving}
								class="inline-flex items-center gap-2 rounded-lg bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-700 disabled:opacity-50 transition-colors"
							>
								<Save size={14} />
								{saving ? m.saving() : m.save()}
							</button>
							<button
								onclick={cancelEdit}
								class="inline-flex items-center gap-1.5 rounded-lg border border-slate-200 px-4 py-2 text-sm text-slate-600 transition-colors hover:bg-slate-50"
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
								class="h-24 w-24 rounded-full object-cover ring-2 ring-slate-100"
							/>
						{:else}
							<div class="flex h-24 w-24 items-center justify-center rounded-full bg-slate-100 ring-2 ring-slate-100">
								<User size={40} class="text-slate-400" />
							</div>
						{/if}
					</div>

					<div class="flex-1 space-y-3">
						<div>
							<p class="text-xs font-medium text-slate-400 mb-1">{m.nickname_label()}</p>
							<p class="text-sm text-slate-800">{profile?.nickname || $user.username}</p>
						</div>
						<div>
							<p class="text-xs font-medium text-slate-400 mb-1">{m.bio_label()}</p>
							<p class="text-sm text-slate-800 whitespace-pre-wrap">{profile?.bio || m.no_bio()}</p>
						</div>
					</div>
				</div>
			{/if}
		</div>

		<!-- Account info -->
		<div class="rounded-lg border bg-white p-6">
			<h2 class="mb-4 text-sm font-medium text-slate-500">{m.account_info()}</h2>
			<div class="space-y-2 text-sm">
				<div class="flex items-center gap-2">
					<span class="text-slate-500">{m.username_label()}</span>
					<span class="text-slate-800">{$user.username}</span>
				</div>
				<div class="flex items-center gap-2">
					<span class="text-slate-500">{m.email_label()}</span>
					<span class="text-slate-800">{$user.email}</span>
				</div>
			</div>
		</div>

		<!-- Storage -->
		<div class="rounded-lg border bg-white p-6">
			<h2 class="mb-4 flex items-center gap-2 text-sm font-medium text-slate-500">
				<HardDrive size={16} /> {m.drive_storage()}
			</h2>
			<div class="space-y-3">
				{#if loading}
					<p class="text-sm text-slate-400">{m.loading()}</p>
				{:else if stats}
					<div class="flex items-baseline gap-2">
						<span class="text-2xl font-semibold text-slate-900">{fmtSize(stats.used_bytes)}</span>
						<span class="text-sm text-slate-500">/ {fmtSize(stats.total_bytes)}</span>
					</div>
					<div class="h-2 w-full overflow-hidden rounded-full bg-slate-100">
						<div
							class="h-full rounded-full bg-slate-900 transition-all"
							style="width:{Math.min((stats.used_bytes / stats.total_bytes) * 100, 100)}%"
						></div>
					</div>
					<div class="flex flex-wrap gap-x-4 gap-y-1 text-xs text-slate-400">
						<span>{m.storage_base()}: {fmtSize(stats.base_bytes)}</span>
						<span>{m.storage_bonus()}: {fmtSize(stats.member_bonus_bytes)}</span>
						<span>{m.storage_pack()}: {fmtSize(stats.pack_bytes)}</span>
					</div>
				{:else}
					<p class="text-sm text-red-600">{m.load_failed()}</p>
				{/if}
			</div>
		</div>
	</div>
{:else}
	<p class="text-slate-600">{@html m.please_login({ link: '<a href="/login" class="underline">' + m.login_link_text() + '</a>' })}</p>
{/if}
