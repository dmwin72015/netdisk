<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import {
    Bell,
    Film,
    Folder,
    HardDrive,
    Image as ImageIcon,
    ListRestart,
    LogOut,
    Search,
    Settings,
    Share2,
    Shield,
    Star,
    Trash2,
    User,
  } from "@lucide/svelte";
  import { logout } from "$lib/api/auth";
  import { setUser, user } from "$lib/stores/auth";
  import { Dropdown, DropdownBase } from "$lib/ui/dropdown";
  import FileSearchDialog from "$lib/components/FileSearchDialog.svelte";
  import { getFilesUrl } from "$lib/stores/last-section-url";
  import * as m from "$lib/paraglide/messages";

  let { children }: { children: import("svelte").Snippet } = $props();

  let currentPath = $derived(page.url.pathname);
  let accountOpen = $state(false);
  let searchOpen = $state(false);

  function handleSearchKeydown(e: KeyboardEvent) {
    if ((e.ctrlKey || e.metaKey) && e.key === "k") {
      e.preventDefault();
      searchOpen = true;
    }
  }

  type NavItem = {
    href: string;
    match: string;
    label: string;
    icon: typeof Folder;
  };
  type MoreItem = {
    href: string;
    match: string;
    label: string;
    icon: typeof Star;
  };

  const primaryNav: NavItem[] = [
    { href: "/files/all", match: "/files", label: m.nav_files(), icon: Folder },
    {
      href: "/photos",
      match: "/photos",
      label: m.nav_photos(),
      icon: ImageIcon,
    },
    { href: "/media", match: "/media", label: m.nav_media(), icon: Film },
  ];

  const moreItems: MoreItem[] = [
    {
      href: "/files/starred",
      match: "/files/starred",
      label: m.nav_starred(),
      icon: Star,
    },
    {
      href: "/files/trash",
      match: "/files/trash",
      label: m.nav_trash(),
      icon: Trash2,
    },
    { href: "/shares", match: "/shares", label: m.nav_shares(), icon: Share2 },
    {
      href: "/tasks",
      match: "/tasks",
      label: m.upload_title(),
      icon: ListRestart,
    },
    {
      href: "/settings",
      match: "/settings",
      label: m.nav_settings(),
      icon: Settings,
    },
    { href: "/profile", match: "/profile", label: m.nav_account(), icon: User },
  ];

  let isAccountPage = $derived(
    ["/profile", "/settings", "/security-log", "/account-manage"].some((p) =>
      currentPath.startsWith(p),
    ),
  );

  let pageTitle = $derived.by(() => {
    const all = [...primaryNav, ...moreItems];
    const active = all.find((item) => currentPath.startsWith(item.match));
    if (active) return active.label;
    if (currentPath.startsWith("/admin")) return m.admin_panel();
    return m.app_name();
  });

  function isActive(match: string) {
    return currentPath.startsWith(match);
  }

  // 点击「文件」tab 时跳到上次离开的子目录（仅普通点击，保留 ⌘/Ctrl+click 等原生行为）。
  function onFilesClick(event: MouseEvent) {
    if (event.defaultPrevented) return;
    if (event.button !== 0) return;
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey)
      return;
    const target = getFilesUrl();
    if (currentPath === target) return;
    event.preventDefault();
    void goto(target);
  }

  async function handleLogout() {
    await logout();
    setUser(null);
    await goto("/login");
  }
</script>

<svelte:window onkeydown={handleSearchKeydown} />

<div class="bg-surface-muted text-ink flex h-screen">
  <!-- Sidebar -->
  <aside
    class="border-line bg-surface sticky top-0 flex h-screen w-16 shrink-0 flex-col items-center border-r py-4"
  >
    <!-- Logo (flat at rest — no glow) -->
    <a
      href="/"
      class="mb-6 flex items-center justify-center"
      aria-label="NetDisk"
    >
      <span
        class="bg-primary text-primary-on flex h-9 w-9 items-center justify-center rounded-xl"
      >
        <HardDrive size={18} strokeWidth={2} />
      </span>
    </a>

    <!-- Primary navigation (icon + text below) -->
    <nav class="flex w-full flex-col items-center gap-1 px-2">
      {#each primaryNav as item (item.href)}
        {@const Icon = item.icon}
        {@const active = isActive(item.match)}
        <a
          href={item.href}
          onclick={item.match === "/files" ? onFilesClick : undefined}
          aria-current={active ? "page" : undefined}
          class="group relative flex w-full flex-col items-center gap-1 rounded-lg py-2 text-[11px] transition-colors duration-150 {active
            ? 'bg-primary-soft text-primary font-medium'
            : 'text-ink-3 hover:bg-surface-sunken hover:text-ink'}"
        >
          <Icon size={20} strokeWidth={active ? 2.25 : 1.75} />
          <span class="leading-none">{item.label}</span>
        </a>
      {/each}
    </nav>

    <!-- User menu at bottom -->
    <Dropdown
      bind:open={accountOpen}
      triggerClass="mt-auto flex w-[52px] flex-col items-center gap-1 rounded-lg py-2 text-[11px] text-ink-3 transition-colors duration-150 hover:bg-surface-sunken hover:text-ink data-[state=open]:bg-surface-sunken data-[state=open]:text-ink"
      contentClass="min-w-[200px] mb-2"
      sideOffset={8}
      align="start"
    >
      {#snippet trigger()}
        {#if $user?.profile?.avatarUrl}
          <img
            src={$user.profile.avatarUrl}
            alt=""
            loading="lazy"
            class="h-7 w-7 rounded-full object-cover"
          />
        {:else}
          <span
            class="bg-primary-soft text-primary flex h-7 w-7 items-center justify-center rounded-full"
          >
            <User size={14} strokeWidth={2} />
          </span>
        {/if}
        <span class="max-w-13 truncate leading-none">
          {$user?.profile?.displayName || $user?.username || ""}
        </span>
      {/snippet}

      <div class="border-line-soft border-b px-3 py-2 mb-1">
        <p
          class="text-ink truncate text-sm font-medium"
          title={$user?.profile?.displayName || $user?.username || ""}
        >
          {$user?.profile?.displayName || $user?.username || m.nav_account()}
        </p>
        {#if $user?.email}
          <p class="text-ink-3 max-w-46 truncate text-xs">{$user.email}</p>
        {/if}
      </div>

      {#each moreItems as item (item.href)}
        {@const Icon = item.icon}
        <DropdownBase.Item onSelect={() => goto(item.href)}>
          {#snippet icon()}<Icon size={14} class="text-ink-4" />{/snippet}
          {#snippet children()}{item.label}{/snippet}
        </DropdownBase.Item>
      {/each}

      {#if $user?.role === "admin"}
        <DropdownBase.Item onSelect={() => goto("/admin")}>
          {#snippet icon()}<Shield size={14} class="text-admin" />{/snippet}
          {#snippet children()}{m.admin_panel()}{/snippet}
        </DropdownBase.Item>
      {/if}

      <DropdownBase.Separator />
      <DropdownBase.Item variant="destructive" onSelect={handleLogout}>
        {#snippet icon()}<LogOut size={14} class="text-danger" />{/snippet}
        {#snippet children()}{m.logout()}{/snippet}
      </DropdownBase.Item>
    </Dropdown>
  </aside>

  <!-- Main area -->
  <div class="flex min-w-0 flex-1 flex-col">
    <header class="flex items-center gap-3 px-6 pt-5 pb-2">
      {#if !isAccountPage}
        <button
          type="button"
          onclick={() => {
            searchOpen = true;
          }}
          class="border-line bg-surface-sunken text-ink-3 hover:border-ink-5 hover:bg-surface-muted hover:text-ink-2 hidden h-9 max-w-xl flex-1 items-center gap-2 rounded-lg border px-3 text-sm transition-colors duration-150 md:flex"
        >
          <Search size={15} strokeWidth={1.75} />
          <span class="flex-1 text-left">{m.search_files()}…</span>
          <kbd
            class="border-line bg-surface text-ink-3 rounded border px-1.5 py-0.5 text-[10px] font-medium"
          >
            ⌘K
          </kbd>
        </button>
      {/if}

      <div class="ml-auto flex items-center gap-1">
        <button
          type="button"
          class="text-ink-3 hover:bg-surface-sunken hover:text-ink flex h-8 w-8 items-center justify-center rounded-lg transition-colors duration-150"
          aria-label={m.notifications()}
        >
          <Bell size={16} strokeWidth={1.75} />
        </button>
      </div>
    </header>

    <main class="min-h-0 min-w-0 flex-1 overflow-y-auto">
      {@render children()}
    </main>
  </div>
</div>

<FileSearchDialog bind:open={searchOpen} />
