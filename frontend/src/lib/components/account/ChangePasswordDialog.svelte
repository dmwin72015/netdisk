<script lang="ts">
  import Dialog from "$lib/ui/dialog/Dialog.svelte";
  import { changePassword } from "$lib/api/profile";
  import { toast } from "svelte-sonner";
  import * as m from "$lib/paraglide/messages";

  let {
    open = $bindable(false),
  }: {
    open?: boolean;
  } = $props();

  let oldPassword = $state("");
  let newPassword = $state("");
  let confirmPassword = $state("");
  let busy = $state(false);
  let error = $state("");

  function reset() {
    oldPassword = "";
    newPassword = "";
    confirmPassword = "";
    error = "";
    busy = false;
  }

  function onOpenChangeComplete(isOpen: boolean) {
    if (!isOpen) reset();
  }

  async function handleConfirm() {
    error = "";

    if (!oldPassword) {
      error = m.account_old_password_placeholder();
      return false;
    }
    if (!newPassword) {
      error = m.account_new_password_placeholder();
      return false;
    }
    if (newPassword.length < 6) {
      error = m.account_password_too_short();
      return false;
    }
    if (newPassword !== confirmPassword) {
      error = m.account_password_mismatch();
      return false;
    }
    if (oldPassword === newPassword) {
      error = m.account_password_same();
      return false;
    }

    busy = true;
    try {
      await changePassword(oldPassword, newPassword);
      toast.success(m.account_password_changed());
      open = false;
    } catch (e) {
      error = e instanceof Error ? e.message : m.account_password_change_failed();
    } finally {
      busy = false;
    }
  }
</script>

<Dialog
  bind:open
  title={m.account_change_password()}
  confirmText={busy ? m.saving() : m.confirm()}
  cancelText={m.cancel()}
  {onOpenChangeComplete}
  onConfirm={handleConfirm}
  onCancel={() => { open = false; }}
>
  <div class="space-y-4">
    <div>
      <label for="old-pwd" class="mb-1 block text-sm font-medium text-ink-2">
        {m.account_old_password()}
      </label>
      <input
        id="old-pwd"
        type="password"
        bind:value={oldPassword}
        placeholder={m.account_old_password_placeholder()}
        class="w-full rounded-lg border border-line bg-white px-3 py-2 text-sm text-ink-2 placeholder:text-ink-4 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
      />
    </div>
    <div>
      <label for="new-pwd" class="mb-1 block text-sm font-medium text-ink-2">
        {m.account_new_password()}
      </label>
      <input
        id="new-pwd"
        type="password"
        bind:value={newPassword}
        placeholder={m.account_new_password_placeholder()}
        class="w-full rounded-lg border border-line bg-white px-3 py-2 text-sm text-ink-2 placeholder:text-ink-4 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
      />
    </div>
    <div>
      <label for="confirm-pwd" class="mb-1 block text-sm font-medium text-ink-2">
        {m.account_confirm_password()}
      </label>
      <input
        id="confirm-pwd"
        type="password"
        bind:value={confirmPassword}
        placeholder={m.account_confirm_password_placeholder()}
        class="w-full rounded-lg border border-line bg-white px-3 py-2 text-sm text-ink-2 placeholder:text-ink-4 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
      />
    </div>
    {#if error}
      <p class="text-xs text-danger">{error}</p>
    {/if}
  </div>
</Dialog>
