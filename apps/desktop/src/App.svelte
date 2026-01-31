<script lang="ts">
  import { onMount } from 'svelte';
  import { invoke } from '@tauri-apps/api/core';
  import { register, unregister } from '@tauri-apps/plugin-global-shortcut';
  import { readText, readImage } from '@tauri-apps/plugin-clipboard-manager';
  import { notes, tags, capture } from './lib/stores/notes';
  import { theme } from './lib/stores/theme';
  import NotePopup from './lib/components/NotePopup.svelte';
  import SyncSettings from './lib/components/SyncSettings.svelte';
  import NotesBrowser from './lib/components/NotesBrowser.svelte';
  import type { Origin } from './lib/types';

  let status = $state('Ready');
  let showSettings = $state(false);
  let currentView = $state<'home' | 'browser'>('home');

  async function captureClipboard() {
    console.log('captureClipboard called');

    // If popup is already open with content, save it first
    let currentCapture: { content: string; selectedTags: string[]; origin: Origin; imageData?: string } | null = null;
    capture.subscribe(c => {
      if (c.isOpen && (c.content.trim() || c.imageData)) {
        currentCapture = { content: c.content, selectedTags: c.selectedTags, origin: c.origin, imageData: c.imageData };
      }
    })();

    if (currentCapture) {
      console.log('Auto-saving previous capture');
      await notes.create(
        currentCapture.content,
        currentCapture.selectedTags,
        currentCapture.origin,
        currentCapture.imageData
      );
      capture.close();
    }

    try {
      status = 'Capturing...';

      let text = '';
      let imageData: string | undefined;

      try {
        const clipText = await readText();
        console.log('Clipboard text:', clipText ? `"${clipText.slice(0, 50)}..."` : 'empty');
        if (clipText && clipText.trim()) {
          text = clipText;
        }
      } catch (e) {
        console.log('No text in clipboard:', e);
      }

      if (typeof readImage === 'function') {
        try {
          const image = await readImage();
          if (image) {
            console.log('Found image in clipboard');
            const bytes = await image.rgba();
            const width = image.width();
            const height = image.height();

            const canvas = document.createElement('canvas');
            canvas.width = width;
            canvas.height = height;
            const ctx = canvas.getContext('2d')!;
            const imgData = ctx.createImageData(width, height);
            imgData.data.set(bytes);
            ctx.putImageData(imgData, 0, 0);
            imageData = canvas.toDataURL('image/png');
          }
        } catch (e) {
          console.log('No image in clipboard or image read failed:', e);
        }
      }

      if (!text && !imageData) {
        console.log('Nothing found in clipboard');
        status = 'Nothing in clipboard';
        setTimeout(() => status = 'Ready', 2000);
        return;
      }

      let origin: Origin;
      try {
        origin = await invoke<Origin>('detect_origin');
        console.log('Detected origin:', origin);
      } catch (e) {
        console.log('Origin detection failed:', e);
        origin = { type: 'unknown' };
      }

      console.log('Opening capture popup with text length:', text.length, 'image:', !!imageData);
      capture.open(text, origin, imageData);
      status = 'Ready';
    } catch (e) {
      console.error('Capture failed:', e);
      status = 'Capture failed';
      setTimeout(() => status = 'Ready', 2000);
    }
  }

  onMount(async () => {
    await Promise.all([notes.load(), tags.load()]);

    try {
      await register('CommandOrControl+Shift+G', async (event) => {
        console.log('Shortcut triggered:', event.state);
        if (event.state === 'Pressed') {
          await captureClipboard();
        }
      });
      console.log('Global shortcut registered: Cmd+Shift+G');
    } catch (e) {
      console.error('Failed to register shortcut:', e);
      status = 'Shortcut failed';
    }

    return async () => {
      try {
        await unregister('CommandOrControl+Shift+G');
      } catch {
        // Ignore
      }
    };
  });

  function toggleTheme() {
    theme.toggle();
  }
</script>

<main class:dark={$theme === 'dark'}>
  {#if currentView === 'browser'}
    <NotesBrowser onBack={() => currentView = 'home'} />
  {:else}
    <header>
      <button class="view-all-link" onclick={() => currentView = 'browser'}>View All</button>
      <div class="header-right">
        <span class="status">{status}</span>
        <button class="icon-btn" onclick={toggleTheme} aria-label="Toggle theme">
          {$theme === 'dark' ? '☀' : '☾'}
        </button>
        <button class="icon-btn" onclick={() => showSettings = !showSettings} aria-label="Settings">
          {showSettings ? '✕' : '⚙'}
        </button>
      </div>
    </header>

    {#if showSettings}
      <section class="settings-section">
        <SyncSettings />
      </section>
    {/if}

    <section class="capture-card">
      <p>Press <kbd>Cmd</kbd> + <kbd>Shift</kbd> + <kbd>G</kbd> to capture</p>
      <button class="capture-btn" onclick={captureClipboard}>
        Capture from clipboard
      </button>
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
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  }

  main {
    min-height: 100vh;
    padding: 1.5rem;
    max-width: 600px;
    margin: 0 auto;
    background: #f0f0f0;
    color: #222;
    display: flex;
    flex-direction: column;
  }

  main.dark {
    background: #1a1a2e;
    color: #eee;
  }

  header {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 0.75rem;
    margin-bottom: 1.5rem;
  }

  .view-all-link {
    background: none;
    border: none;
    color: #0066cc;
    font-size: 1rem;
    cursor: pointer;
    padding: 0;
    margin-right: auto;
  }

  main.dark .view-all-link {
    color: #6ab0ff;
  }

  .view-all-link:hover {
    text-decoration: underline;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .status {
    font-size: 0.75rem;
    color: #888;
    padding: 0.25rem 0.75rem;
    background: rgba(0,0,0,0.08);
    border-radius: 1rem;
  }

  main.dark .status {
    background: rgba(255,255,255,0.1);
  }

  .icon-btn {
    width: 2rem;
    height: 2rem;
    border-radius: 50%;
    border: 1px solid #ccc;
    background: transparent;
    color: inherit;
    font-size: 1rem;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  main.dark .icon-btn {
    border-color: #555;
  }

  .icon-btn:hover {
    background: rgba(0,0,0,0.05);
  }

  main.dark .icon-btn:hover {
    background: rgba(255,255,255,0.1);
  }

  .settings-section {
    margin-bottom: 1.5rem;
  }

  .capture-card {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background: #e0e0e0;
    border-radius: 12px;
    padding: 3rem;
    min-height: 300px;
  }

  main.dark .capture-card {
    background: #252538;
  }

  .capture-card p {
    margin: 0 0 1.5rem 0;
    color: #666;
    font-size: 1rem;
  }

  main.dark .capture-card p {
    color: #999;
  }

  kbd {
    display: inline-block;
    padding: 0.2rem 0.5rem;
    background: #ccc;
    border-radius: 4px;
    font-family: inherit;
    font-size: 0.9rem;
  }

  main.dark kbd {
    background: #444;
  }

  .capture-btn {
    padding: 0.75rem 1.5rem;
    background: #333;
    border: none;
    border-radius: 8px;
    color: #fff;
    font-size: 1rem;
    cursor: pointer;
  }

  main.dark .capture-btn {
    background: #555;
  }

  .capture-btn:hover {
    opacity: 0.9;
  }
</style>
