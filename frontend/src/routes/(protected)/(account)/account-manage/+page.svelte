<script lang="ts">
  import { onMount } from "svelte";
  import { authReady, user, setUser } from "$lib/stores/auth";
  import { goto } from "$app/navigation";
  import {
    getProfile,
    unlinkOAuth,
    deleteAccount,
    type ProfileData,
    type OAuthAccountInfo,
  } from "$lib/api/profile";
  import { confirmAction, promptInput } from "$lib/dialog";
  import { toast } from "svelte-sonner";
  import { LoaderCircle, AlertTriangle, Check, Link2, Unlink } from "@lucide/svelte";
  import ChangePasswordDialog from "$lib/components/account/ChangePasswordDialog.svelte";
  import * as m from "$lib/paraglide/messages";

  let profile = $state<ProfileData | null>(null);
  let loading = $state(true);

  const providerLabels: Record<string, string> = {
    github: "GitHub",
    "2libra": "2Libra",
  };

  async function fetchProfile() {
    loading = true;
    try {
      profile = await getProfile();
      if (!profile.oauthAccounts) profile.oauthAccounts = [];
    } catch {
      // ignore
    } finally {
      loading = false;
    }
  }

  function hasOAuth(provider: string): boolean {
    return (profile?.oauthAccounts ?? []).some((a) => a.provider === provider);
  }

  function getOAuth(provider: string): OAuthAccountInfo | undefined {
    return (profile?.oauthAccounts ?? []).find((a) => a.provider === provider);
  }

  async function handleUnlink(provider: string) {
    const ok = await confirmAction(
      m.account_unlink_title(),
      m.account_unlink_confirm({ provider: providerLabels[provider] || provider }),
      m.confirm(),
    );
    if (!ok) return;
    try {
      await unlinkOAuth(provider);
      toast.success(m.account_unlink_success());
      await fetchProfile();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.account_unlink_failed());
    }
  }

  async function handleDelete() {
    const hint = m.account_delete_prompt_hint();
    const input = await promptInput(
      m.account_delete(),
      m.account_delete_prompt({ text: hint }),
      "",
      hint.length,
    );
    if (!input) return;
    if (input !== hint) {
      toast.error(m.account_delete_failed());
      return;
    }
    const ok = await confirmAction(
      m.account_delete_confirm_title(),
      m.account_delete_confirm_desc(),
      m.account_delete(),
    );
    if (!ok) return;
    try {
      await deleteAccount();
      toast.success(m.account_delete_success());
      goto("/login");
    } catch (e) {
      toast.error(e instanceof Error ? e.message : m.account_delete_failed());
    }
  }

  let showChangePassword = $state(false);

  let securityLevel = $derived.by(() => {
    if (!profile) return { level: 0, tips: [] as string[] };
    let level = 0;
    const tips: string[] = [];
    if (profile.email) level++;
    else tips.push(m.account_security_no_email());
    if (profile.oauthAccounts?.length > 0) level++;
    else tips.push(m.account_security_no_oauth());
    return { level, tips };
  });

  onMount(() => {
    void fetchProfile();
  });
</script>

{#if $authReady && $user}
  <div class="space-y-8">
    {#if loading}
      <div class="flex items-center justify-center py-24">
        <LoaderCircle size={20} class="animate-spin text-ink-4" />
      </div>
    {:else if profile}
      <!-- Security warning (暂时隐藏) -->
      {#if false && securityLevel.level < 2}
        <div class="flex items-start gap-2 rounded-lg border border-warning/30 bg-warning-soft px-4 py-3">
          <AlertTriangle size={16} class="mt-0.5 shrink-0 text-warning" />
          <span class="text-sm text-ink-2">{m.account_security_low()}</span>
        </div>
      {/if}

      <!-- Account binding -->
      <section>
        <h2 class="mb-4 text-sm font-semibold text-ink">{m.account_binding()}</h2>
        <div class="space-y-0 divide-y divide-line-soft rounded-lg border border-line-soft bg-surface">
          <!-- Email -->
          <div class="flex items-center gap-3 px-4 py-3">
            {#if profile.email}
              <Check size={16} class="shrink-0 text-success" />
            {:else}
              <AlertTriangle size={16} class="shrink-0 text-warning" />
            {/if}
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-ink-2">{m.account_email()}</p>
              <p class="text-xs text-ink-4">
                {profile.email || m.account_email_not_bound()}
              </p>
            </div>
          </div>

          <!-- Password -->
          <div class="flex items-center gap-3 px-4 py-3">
            <Check size={16} class="shrink-0 text-success" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-ink-2">{m.account_password()}</p>
              <p class="text-xs text-ink-4">{m.account_password_set()}</p>
            </div>
            <button
              type="button"
              onclick={() => { showChangePassword = true; }}
              class="shrink-0 rounded-md px-3 py-1 text-xs font-medium text-ink-3 transition-colors hover:bg-surface-sunken hover:text-ink"
            >
              {m.account_change()}
            </button>
          </div>
        </div>
      </section>

      <!-- Third-party accounts -->
      <section>
        <h2 class="mb-1 text-sm font-semibold text-ink">{m.account_third_party()}</h2>
        <p class="mb-4 text-xs text-ink-4">{m.account_third_party_desc()}</p>
        <div class="space-y-0 divide-y divide-line-soft rounded-lg border border-line-soft bg-surface">
          {#each ["github", "2libra"] as provider}
            {@const bound = hasOAuth(provider)}
            {@const info = getOAuth(provider)}
            <div class="flex items-center gap-3 px-4 py-3">
              <Link2 size={16} class="shrink-0 text-ink-3" />
              <div class="min-w-0 flex-1">
                <p class="text-sm font-medium text-ink-2">
                  {providerLabels[provider] || provider}
                </p>
                <p class="text-xs text-ink-4">
                  {#if bound && info}
                    {info.oauthEmail || info.providerAccountId}
                  {:else}
                    {m.account_not_bound()}
                  {/if}
                </p>
              </div>
              {#if bound}
                <button
                  type="button"
                  onclick={() => handleUnlink(provider)}
                  class="shrink-0 rounded-md px-3 py-1 text-xs font-medium text-danger transition-colors hover:bg-danger-soft"
                >
                  {m.account_unlink()}
                </button>
              {:else}
                <a
                  href="/api/v1/auth/oauth/{provider}/authorize"
                  class="shrink-0 rounded-md px-3 py-1 text-xs font-medium text-primary transition-colors hover:bg-primary-soft"
                >
                  {m.account_bind()}
                </a>
              {/if}
            </div>
          {/each}
        </div>
      </section>

      <!-- Delete account -->
      <section>
        <h2 class="mb-4 text-sm font-semibold text-ink">{m.account_delete()}</h2>
        <div class="rounded-lg border border-line-soft bg-surface px-4 py-4">
          <p class="text-sm text-ink-3">{m.account_delete_desc()}</p>
          <button
            type="button"
            onclick={handleDelete}
            class="mt-3 rounded-lg bg-danger px-4 py-2 text-sm font-medium text-primary-on transition-colors hover:bg-danger-hover"
          >
            {m.account_delete()}
          </button>
        </div>
      </section>
    {/if}
  </div>

  <ChangePasswordDialog bind:open={showChangePassword} />
{/if}
