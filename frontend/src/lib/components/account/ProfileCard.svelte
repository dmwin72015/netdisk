<script lang="ts">
	import { Camera, Save, User, Pencil, X } from '@lucide/svelte';
	import { updateProfile, uploadAvatar } from '$lib/api/profile';
	import * as m from '$lib/paraglide/messages';
	import { getAvatarMaxSize, configError } from '$lib/stores/config';

	let {
		displayName,
		avatarUrl,
		bio,
		username,
		onSaved
	}: {
		displayName: string;
		avatarUrl: string;
		bio: string;
		username: string;
		onSaved: () => void;
	} = $props();

	let editing = $state(false);
	let saving = $state(false);
	let saveMsg = $state('');
	let nickname = $state('');
	let editBio = $state('');
	let avatarPreview = $state<string | null>(null);
	let avatarFile = $state<File | null>(null);

	function startEdit() {
		nickname = displayName;
		editBio = bio;
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
		if ($configError) {
			saveMsg = m.config_unavailable();
			input.value = '';
			return;
		}
		const maxSize = getAvatarMaxSize();
		if (maxSize === null || file.size > maxSize) {
			saveMsg = m.upload_failed();
			input.value = '';
			return;
		}
		avatarFile = file;
		const reader = new FileReader();
		reader.onload = () => (avatarPreview = reader.result as string);
		reader.readAsDataURL(file);
	}

	async function handleSave() {
		saving = true;
		saveMsg = '';
		try {
			let newAvatarUrl: string | undefined;
			if (avatarFile) {
				newAvatarUrl = await uploadAvatar(avatarFile);
				avatarFile = null;
			}
			await updateProfile({ displayName: nickname, bio: editBio, avatarUrl: newAvatarUrl });
			saveMsg = m.profile_saved();
			editing = false;
			avatarPreview = null;
			onSaved();
		} catch {
			saveMsg = m.profile_save_failed();
		} finally {
			saving = false;
		}
	}
</script>

<div class="rounded-xl border border-line-soft bg-surface p-6 ">
	<div class="mb-4 flex items-center justify-between">
		<h2 class="text-sm font-medium text-ink-3">{m.profile_info()}</h2>
		{#if !editing}
			<button
				onclick={startEdit}
				class="inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm text-ink-3 transition-colors hover:bg-surface-sunken hover:text-ink"
			>
				<Pencil size={14} />
				{m.edit()}
			</button>
		{/if}
	</div>

	{#if editing}
		<div class="flex flex-col gap-6 sm:flex-row">
			<div class="flex flex-col items-center gap-3">
				<div class="relative">
					{#if avatarPreview || avatarUrl}
							<img
								src={avatarPreview || avatarUrl}
								alt="avatar"
								loading="lazy"
								class="h-24 w-24 rounded-full object-cover ring-2 ring-line-soft"
							/>
					{:else}
						<div class="flex h-24 w-24 items-center justify-center rounded-full bg-surface-sunken ring-2 ring-line-soft">
							<User size={40} class="text-ink-4" />
						</div>
					{/if}
					<label
						class="absolute -bottom-1 -right-1 flex h-8 w-8 cursor-pointer items-center justify-center rounded-full bg-ink text-primary-on hover:bg-ink-2 transition-colors"
					>
						<Camera size={14} />
						<input type="file" accept="image/jpeg,image/png,image/webp" class="hidden" onchange={onAvatarSelect} />
					</label>
				</div>
				<p class="text-xs text-ink-4">{m.avatar_hint()}</p>
			</div>

			<div class="flex-1 space-y-4">
				<div>
					<label class="mb-1 block text-sm font-medium text-ink-2" for="nickname">{m.nickname_label()}</label>
					<input
						id="nickname"
						type="text"
						bind:value={nickname}
						placeholder={username}
						maxlength={100}
						class="w-full rounded-lg border border-line px-3 py-2 text-sm focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none"
					/>
				</div>
				<div>
					<label class="mb-1 block text-sm font-medium text-ink-2" for="bio">{m.bio_label()}</label>
					<textarea
						id="bio"
						bind:value={editBio}
						placeholder={m.bio_placeholder()}
						rows={3}
						maxlength={500}
						class="w-full rounded-lg border border-line px-3 py-2 text-sm focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none resize-none"
					></textarea>
				</div>

				<div class="flex items-center gap-3">
					<button
						onclick={handleSave}
						disabled={saving}
						class="inline-flex items-center gap-2 rounded-lg bg-ink px-4 py-2 text-sm font-medium text-primary-on hover:bg-ink-2 disabled:opacity-50 transition-colors"
					>
						<Save size={14} />
						{saving ? m.saving() : m.save()}
					</button>
					<button
						onclick={cancelEdit}
						class="inline-flex items-center gap-1.5 rounded-lg border border-line px-4 py-2 text-sm text-ink-3 transition-colors hover:bg-surface-sunken"
					>
						<X size={14} />
						{m.cancel()}
					</button>
					{#if saveMsg}
						<span class="text-sm {saveMsg === m.profile_saved() ? 'text-success' : 'text-danger'}">{saveMsg}</span>
					{/if}
				</div>
			</div>
		</div>
	{:else}
		<div class="flex flex-col gap-6 sm:flex-row">
			<div class="flex flex-col items-center gap-3">
				{#if avatarUrl}
						<img
							src={avatarUrl}
							alt="avatar"
							loading="lazy"
							class="h-24 w-24 rounded-full object-cover ring-2 ring-line-soft"
						/>
				{:else}
					<div class="flex h-24 w-24 items-center justify-center rounded-full bg-surface-sunken ring-2 ring-line-soft">
						<User size={40} class="text-ink-4" />
					</div>
				{/if}
			</div>

			<div class="flex-1 space-y-3">
				<div>
					<p class="text-xs font-medium text-ink-4 mb-1">{m.nickname_label()}</p>
					<p class="text-sm text-ink-2">{displayName || username}</p>
				</div>
				<div>
					<p class="text-xs font-medium text-ink-4 mb-1">{m.bio_label()}</p>
					<p class="text-sm text-ink-2 whitespace-pre-wrap">{bio || m.no_bio()}</p>
				</div>
			</div>
		</div>
	{/if}
</div>
