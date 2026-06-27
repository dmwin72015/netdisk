<script lang="ts">
  import { authReady, user } from "$lib/stores/auth";
  import { UserCog } from "@lucide/svelte";
  import * as m from "$lib/paraglide/messages";
</script>

{#if $authReady && $user}
  <div class="space-y-6">
    <h1 class="text-xl font-semibold text-ink">{m.nav_account_manage()}</h1>
    <div class="flex flex-col items-center justify-center py-24 text-center">
      <UserCog size={48} class="mb-4 text-ink-4" />
      <p class="text-sm text-ink-4">{m.coming_soon()}</p>
    </div>
  </div>
{/if}
