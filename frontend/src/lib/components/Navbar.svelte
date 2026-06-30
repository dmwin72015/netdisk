<script lang="ts">
	import { user, authReady } from '$lib/stores/auth';
	import { logout } from '$lib/api/auth';
	import { goto } from '$app/navigation';
	import { setUser } from '$lib/stores/auth';
	import { HardDrive, Folder, Star, Trash2, Film, Image as ImageIcon, User, ChevronDown, LogOut, Settings, ListRestart, Shield } from '@lucide/svelte';
	import { Dropdown, DropdownBase } from '$lib/ui/dropdown';
	import LanguageDropdown from '$lib/components/LanguageDropdown.svelte';
	import { page } from '$app/state';
	import * as m from '$lib/paraglide/messages';

	async function handleLogout() {
		await logout();
		setUser(null);
		await goto('/login');
	}

	let currentPath = $derived(page.url.pathname);

	function isActive(path: string): boolean {
		if (path === '/') return currentPath === '/';
		return currentPath.startsWith(path);
	}

	let accountOpen = $state(false);
</script>

<header class="sticky top-0 z-40 border-b border-line bg-surface/80 backdrop-blur-sm">
	<div class="mx-auto flex h-14 max-w-6xl items-center justify-between px-4">
		<a href="/" class="flex items-center gap-2 font-semibold text-ink transition-colors hover:text-primary">
			<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-on">
				<HardDrive size={18} />
			</div>
			<span class="text-base">Netdisk</span>
		</a>
		<nav class="flex items-center gap-1 text-sm">
			{#if $authReady}
				{#if $user}
					<a
						href="/files/all"
						class="flex items-center gap-1.5 rounded-lg px-3 py-1.5 transition-colors {isActive('/files') ? 'bg-primary-soft text-primary font-medium' : 'text-ink-3 hover:bg-surface-sunken hover:text-ink'}"
					>
						<Folder size={15} /> {m.nav_files()}
					</a>
					<a
						href="/media"
						class="flex items-center gap-1.5 rounded-lg px-3 py-1.5 transition-colors {isActive('/media') ? 'bg-primary-soft text-primary font-medium' : 'text-ink-3 hover:bg-surface-sunken hover:text-ink'}"
					>
						<Film size={15} /> {m.nav_media()}
					</a>
					<a
						href="/photos"
						class="flex items-center gap-1.5 rounded-lg px-3 py-1.5 transition-colors {isActive('/photos') ? 'bg-primary-soft text-primary font-medium' : 'text-ink-3 hover:bg-surface-sunken hover:text-ink'}"
					>
						<ImageIcon size={15} /> {m.nav_photos()}
					</a>
					<div class="mx-1 h-5 w-px bg-line"></div>

					<!-- Account dropdown -->
					<Dropdown
						bind:open={accountOpen}
						triggerClass="flex items-center justify-center gap-1 rounded-lg px-2.5 py-1.5 text-ink-3 transition-colors hover:bg-surface-sunken hover:text-ink data-[state=open]:bg-surface-sunken data-[state=open]:text-ink"
						contentClass="min-w-[160px]"
						sideOffset={8}
						align="end"
					>
						{#snippet trigger()}
							{#if $user?.profile?.avatarUrl}
									<img src={$user.profile.avatarUrl} alt="" loading="lazy" class="h-5 w-5 rounded-full object-cover" />
							{:else}
								<User size={15} />
							{/if}
							<span
								class="hidden max-w-[10rem] truncate sm:inline-block"
								title={$user?.profile?.displayName || $user?.username || m.nav_account()}
							>
								{$user?.profile?.displayName || $user?.username || m.nav_account()}
							</span>
							<ChevronDown size={12} class="text-ink-4" />
						{/snippet}

						<DropdownBase.Item onSelect={() => goto('/files/starred')}>
							{#snippet icon()}<Star size={14} class="text-ink-4" />{/snippet}
							{#snippet children()}{m.nav_starred()}{/snippet}
						</DropdownBase.Item>
						<DropdownBase.Item onSelect={() => goto('/files/trash')}>
							{#snippet icon()}<Trash2 size={14} class="text-ink-4" />{/snippet}
							{#snippet children()}{m.nav_trash()}{/snippet}
						</DropdownBase.Item>
						<DropdownBase.Separator />
						{#if $user.role === 'admin'}
							<DropdownBase.Item onSelect={() => goto('/admin')}>
								{#snippet icon()}<Shield size={14} class="text-warning" />{/snippet}
								{#snippet children()}{m.admin_panel()}{/snippet}
							</DropdownBase.Item>
							<DropdownBase.Separator />
						{/if}
						<DropdownBase.Item onSelect={() => goto('/tasks')}>
							{#snippet icon()}<ListRestart size={14} class="text-ink-4" />{/snippet}
							{#snippet children()}{m.upload_title()}{/snippet}
						</DropdownBase.Item>
						<DropdownBase.Item onSelect={() => goto('/account')}>
							{#snippet icon()}<Settings size={14} class="text-ink-4" />{/snippet}
							{#snippet children()}{m.nav_account()}{/snippet}
						</DropdownBase.Item>
						<DropdownBase.Separator />
						<DropdownBase.Item variant="destructive" onSelect={handleLogout}>
							{#snippet icon()}<LogOut size={14} class="text-danger" />{/snippet}
							{#snippet children()}{m.logout()}{/snippet}
						</DropdownBase.Item>
					</Dropdown>
				{:else}
					<a href="/login" class="rounded-lg px-3 py-1.5 text-ink-3 transition-colors hover:bg-surface-sunken hover:text-ink">{m.nav_login()}</a>
					<a href="/register" class="rounded-lg bg-primary px-3 py-1.5 text-primary-on transition-colors hover:bg-primary-hover">{m.nav_register()}</a>
				{/if}
			{/if}

			<div class="mx-1 h-5 w-px bg-line"></div>

			<!-- Language dropdown -->
			<LanguageDropdown />
		</nav>
	</div>
</header>
