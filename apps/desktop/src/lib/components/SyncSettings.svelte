<script lang="ts">
  import { syncStore } from '../stores/sync';
  import { onMount } from 'svelte';

  let serverUrl = $state('http://localhost:8080');
  let email = $state('');
  let password = $state('');
  let isRegistering = $state(false);
  let isLoading = $state(false);
  let message = $state('');

  onMount(() => {
    syncStore.refreshStatus();
  });

  async function toggleSync() {
    isLoading = true;
    message = '';
    try {
      const newEnabled = !$syncStore.status.enabled;
      await syncStore.configure(serverUrl, newEnabled);
      message = newEnabled ? 'Sync enabled' : 'Sync disabled';
    } catch (e) {
      message = `Error: ${e}`;
    }
    isLoading = false;
  }

  async function handleAuth() {
    if (!email || !password) {
      message = 'Please enter email and password';
      return;
    }

    isLoading = true;
    message = '';
    try {
      if (isRegistering) {
        await syncStore.register(email, password);
        message = 'Account created successfully';
      } else {
        await syncStore.login(email, password);
        message = 'Logged in successfully';
      }
      password = '';
    } catch (e) {
      message = `Error: ${e}`;
    }
    isLoading = false;
  }

  async function handleLogout() {
    await syncStore.logout();
    message = 'Logged out';
    email = '';
    password = '';
  }

  async function handleSync() {
    isLoading = true;
    message = '';
    try {
      await syncStore.syncNow();
      message = 'Sync completed';
    } catch (e) {
      message = `Sync failed: ${e}`;
    }
    isLoading = false;
  }
</script>

<div class="sync-settings">
  <h3>Sync Settings</h3>

  <div class="setting-row">
    <label for="server-url">Server URL</label>
    <input
      id="server-url"
      type="text"
      bind:value={serverUrl}
      placeholder="http://localhost:8080"
      disabled={$syncStore.status.enabled}
    />
  </div>

  <div class="setting-row">
    <span>Sync Status</span>
    <button
      class="toggle"
      class:enabled={$syncStore.status.enabled}
      onclick={toggleSync}
      disabled={isLoading}
    >
      {$syncStore.status.enabled ? 'Enabled' : 'Disabled'}
    </button>
  </div>

  {#if $syncStore.status.enabled}
    <hr />

    {#if !$syncStore.status.authenticated}
      <div class="auth-section">
        <div class="tabs">
          <button
            class:active={!isRegistering}
            onclick={() => isRegistering = false}
          >
            Login
          </button>
          <button
            class:active={isRegistering}
            onclick={() => isRegistering = true}
          >
            Register
          </button>
        </div>

        <div class="form">
          <input
            type="email"
            bind:value={email}
            placeholder="Email"
            disabled={isLoading}
          />
          <input
            type="password"
            bind:value={password}
            placeholder="Password"
            disabled={isLoading}
            onkeydown={(e) => e.key === 'Enter' && handleAuth()}
          />
          <button
            class="primary"
            onclick={handleAuth}
            disabled={isLoading}
          >
            {isLoading ? 'Loading...' : isRegistering ? 'Create Account' : 'Login'}
          </button>
        </div>
      </div>
    {:else}
      <div class="logged-in">
        <p>Logged in</p>
        <div class="actions">
          <button
            class="primary"
            onclick={handleSync}
            disabled={$syncStore.isSyncing}
          >
            {$syncStore.isSyncing ? 'Syncing...' : 'Sync Now'}
          </button>
          <button onclick={handleLogout}>Logout</button>
        </div>
      </div>
    {/if}
  {/if}

  {#if message}
    <p class="message" class:error={message.startsWith('Error') || message.includes('failed')}>
      {message}
    </p>
  {/if}
</div>

<style>
  .sync-settings {
    padding: 1rem;
    background: #252538;
    border-radius: 8px;
  }

  h3 {
    margin: 0 0 1rem 0;
    font-size: 1rem;
    color: #fff;
  }

  .setting-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 0.75rem;
  }

  .setting-row label,
  .setting-row span {
    color: #aaa;
    font-size: 0.875rem;
  }

  input[type="text"],
  input[type="email"],
  input[type="password"] {
    padding: 0.5rem;
    border-radius: 4px;
    border: 1px solid #444;
    background: #1a1a2e;
    color: #fff;
    font-size: 0.875rem;
    width: 200px;
  }

  input:focus {
    outline: none;
    border-color: #4a9eff;
  }

  input:disabled {
    opacity: 0.6;
  }

  .toggle {
    padding: 0.375rem 0.75rem;
    border-radius: 1rem;
    border: 1px solid #444;
    background: transparent;
    color: #888;
    font-size: 0.8rem;
    cursor: pointer;
  }

  .toggle.enabled {
    background: #2d5a3d;
    border-color: #3d7a4d;
    color: #8f8;
  }

  hr {
    border: none;
    border-top: 1px solid #333;
    margin: 1rem 0;
  }

  .auth-section {
    margin-top: 0.5rem;
  }

  .tabs {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }

  .tabs button {
    flex: 1;
    padding: 0.5rem;
    border: none;
    background: #333;
    color: #888;
    border-radius: 4px;
    cursor: pointer;
  }

  .tabs button.active {
    background: #4a9eff;
    color: #fff;
  }

  .form {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .form input {
    width: 100%;
    box-sizing: border-box;
  }

  button.primary {
    padding: 0.5rem 1rem;
    background: #4a9eff;
    border: none;
    border-radius: 4px;
    color: #fff;
    cursor: pointer;
  }

  button.primary:hover:not(:disabled) {
    background: #3a8eef;
  }

  button.primary:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .logged-in {
    text-align: center;
  }

  .logged-in p {
    color: #8f8;
    margin: 0 0 1rem 0;
  }

  .actions {
    display: flex;
    gap: 0.5rem;
    justify-content: center;
  }

  .actions button:not(.primary) {
    padding: 0.5rem 1rem;
    background: transparent;
    border: 1px solid #444;
    border-radius: 4px;
    color: #aaa;
    cursor: pointer;
  }

  .message {
    margin-top: 1rem;
    padding: 0.5rem;
    border-radius: 4px;
    background: #2d5a3d;
    color: #8f8;
    font-size: 0.8rem;
    text-align: center;
  }

  .message.error {
    background: #5a2d2d;
    color: #f88;
  }
</style>
