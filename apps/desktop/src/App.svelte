<script lang="ts">
  import { onMount } from 'svelte';
  import { invoke } from '@tauri-apps/api/core';
  import { register, unregister } from '@tauri-apps/plugin-global-shortcut';
  import { readText } from '@tauri-apps/plugin-clipboard-manager';
  import { notes, tags, capture } from './lib/stores/notes';
  import NotePopup from './lib/components/NotePopup.svelte';
  import SyncSettings from './lib/components/SyncSettings.svelte';
  import type { Origin, Note } from './lib/types';

  let status = $state('Ready');
  let recentNotes: Note[] = $state([]);
  let showSettings = $state(false);

  // Subscribe to notes store
  notes.subscribe(value => {
    recentNotes = value.slice(0, 5);
  });

  async function captureText() {
    try {
      status = 'Capturing...';

      // Read from clipboard (user should have text selected and we read after copy)
      const text = await readText();

      if (!text || !text.trim()) {
        status = 'No text in clipboard';
        setTimeout(() => status = 'Ready', 2000);
        return;
      }

      // Detect origin (where the text came from)
      let origin: Origin;
      try {
        origin = await invoke<Origin>('detect_origin');
      } catch {
        origin = { type: 'unknown' };
      }

      // Open the capture popup
      capture.open(text, origin);
      status = 'Ready';
    } catch (e) {
      console.error('Capture failed:', e);
      status = 'Capture failed';
      setTimeout(() => status = 'Ready', 2000);
    }
  }

  onMount(async () => {
    // Load existing notes and tags
    await Promise.all([notes.load(), tags.load()]);

    // Register global shortcut
    try {
      await register('CommandOrControl+Shift+G', async (event) => {
        if (event.state === 'Pressed') {
          await captureText();
        }
      });
      console.log('Global shortcut registered: Cmd+Shift+G');
    } catch (e) {
      console.error('Failed to register shortcut:', e);
      status = 'Shortcut failed';
    }

    // Cleanup on unmount
    return async () => {
      try {
        await unregister('CommandOrControl+Shift+G');
      } catch {
        // Ignore cleanup errors
      }
    };
  });

  function formatDate(dateStr: string): string {
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
    });
  }

  function truncate(text: string, maxLength: number): string {
    if (text.length <= maxLength) return text;
    return text.slice(0, maxLength) + '...';
  }
</script>

<main>
  <header>
    <h1>Glean</h1>
    <div class="header-right">
      <span class="status">{status}</span>
      <button class="settings-btn" onclick={() => showSettings = !showSettings} aria-label="Settings">
        {showSettings ? '✕' : '⚙'}
      </button>
    </div>
  </header>

  {#if showSettings}
    <section class="settings-section">
      <SyncSettings />
    </section>
  {/if}

  <section class="shortcut-info">
    <p>Press <kbd>⌘</kbd> + <kbd>⇧</kbd> + <kbd>G</kbd> to capture selected text</p>
    <button class="capture-btn" onclick={captureText}>
      Or click here to capture from clipboard
    </button>
  </section>

  {#if recentNotes.length > 0}
    <section class="recent-notes">
      <h2>Recent Notes</h2>
      <ul>
        {#each recentNotes as note (note.id)}
          <li class="note-item">
            <p class="note-content">{truncate(note.content, 100)}</p>
            <div class="note-meta">
              <span class="note-date">{formatDate(note.created_at)}</span>
              {#if note.tags.length > 0}
                <span class="note-tags">
                  {note.tags.join(', ')}
                </span>
              {/if}
            </div>
          </li>
        {/each}
      </ul>
    </section>
  {:else}
    <section class="empty-state">
      <p>No notes yet. Capture your first text!</p>
    </section>
  {/if}
</main>

<NotePopup />

<style>
  :global(*) {
    box-sizing: border-box;
  }

  :global(body) {
    margin: 0;
    padding: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
    background: #1a1a2e;
    color: #eee;
  }

  main {
    min-height: 100vh;
    padding: 1.5rem;
    max-width: 600px;
    margin: 0 auto;
  }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 2rem;
  }

  h1 {
    font-size: 1.5rem;
    margin: 0;
    color: #fff;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .status {
    font-size: 0.75rem;
    color: #888;
    padding: 0.25rem 0.75rem;
    background: #252538;
    border-radius: 1rem;
  }

  .settings-btn {
    width: 2rem;
    height: 2rem;
    border-radius: 50%;
    border: 1px solid #444;
    background: transparent;
    color: #888;
    font-size: 1rem;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .settings-btn:hover {
    border-color: #666;
    color: #fff;
  }

  .settings-section {
    margin-bottom: 2rem;
  }

  .shortcut-info {
    text-align: center;
    padding: 2rem;
    background: #252538;
    border-radius: 12px;
    margin-bottom: 2rem;
  }

  .shortcut-info p {
    margin: 0 0 1rem 0;
    color: #aaa;
  }

  kbd {
    display: inline-block;
    padding: 0.25rem 0.5rem;
    background: #333;
    border: 1px solid #555;
    border-radius: 4px;
    font-family: inherit;
    font-size: 0.875rem;
    min-width: 1.5rem;
    text-align: center;
  }

  .capture-btn {
    padding: 0.75rem 1.5rem;
    background: #4a9eff;
    border: none;
    border-radius: 8px;
    color: #fff;
    font-size: 0.875rem;
    cursor: pointer;
    transition: background 0.15s ease;
  }

  .capture-btn:hover {
    background: #3a8eef;
  }

  .recent-notes h2 {
    font-size: 0.875rem;
    color: #888;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin: 0 0 1rem 0;
  }

  .recent-notes ul {
    list-style: none;
    padding: 0;
    margin: 0;
  }

  .note-item {
    padding: 1rem;
    background: #252538;
    border-radius: 8px;
    margin-bottom: 0.75rem;
  }

  .note-content {
    margin: 0 0 0.5rem 0;
    color: #ddd;
    font-size: 0.9rem;
    line-height: 1.4;
  }

  .note-meta {
    display: flex;
    align-items: center;
    gap: 1rem;
    font-size: 0.75rem;
    color: #666;
  }

  .note-tags {
    color: #4a9eff;
  }

  .empty-state {
    text-align: center;
    padding: 3rem;
    color: #666;
  }
</style>
