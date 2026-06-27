<script lang="ts">
  import { page } from "$app/state";
  import { User, Settings, Shield, UserCog } from "@lucide/svelte";
  import * as m from "$lib/paraglide/messages";

  let { children } = $props();

  const navItems = [
    { href: "/profile", match: "/profile", label: m.nav_account(), icon: User },
    { href: "/settings", match: "/settings", label: m.nav_settings(), icon: Settings },
    { href: "/security-log", match: "/security-log", label: m.nav_security_log(), icon: Shield },
    { href: "/account-manage", match: "/account-manage", label: m.nav_account_manage(), icon: UserCog },
  ];

  function isActive(match: string) {
    return page.url.pathname.startsWith(match);
  }
</script>

<div class="flex gap-8 px-6 pt-4 pb-6">
  <!-- Left sidebar nav -->
  <nav class="w-48 shrink-0 sticky top-4 self-start">
    <ul class="space-y-0.5">
      {#each navItems as item (item.href)}
        {@const Icon = item.icon}
        {@const active = isActive(item.match)}
        <li>
          <a
            href={item.href}
            class="flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm transition-colors {active
              ? 'bg-primary-soft text-primary font-medium'
              : 'text-ink-3 hover:bg-surface-sunken hover:text-ink'}"
          >
            <Icon size={16} strokeWidth={active ? 2 : 1.75} />
            {item.label}
          </a>
        </li>
      {/each}
    </ul>
  </nav>

  <!-- Right content -->
  <div class="min-w-0 flex-1">
    {@render children()}
  </div>
</div>
