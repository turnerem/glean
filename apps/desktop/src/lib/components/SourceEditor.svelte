<script lang="ts">
  import { openUrl as tauriOpenUrl } from '@tauri-apps/plugin-opener';
  import { theme } from '../stores/theme';
  import type { Origin } from '../types';

  interface Props {
    origin: Origin;
    onOriginChange: (origin: Origin) => void;
  }

  let { origin, onOriginChange }: Props = $props();

  function handleTitleChange(e: Event) {
    const target = e.target as HTMLInputElement;
    onOriginChange({ ...origin, title: target.value || undefined });
  }

  function clearUrl() {
    onOriginChange({ ...origin, url: undefined });
  }

  async function handleUrlClick(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();
    if (origin.url) {
      try {
        await tauriOpenUrl(origin.url);
      } catch (err) {
        console.error('Failed to open URL:', err);
        // Fallback to window.open
        window.open(origin.url, '_blank');
      }
    }
  }
</script>

<div class="source-editor" class:dark={$theme === 'dark'}>
  <input
    type="text"
    class="source-input"
    value={origin.title || ''}
    oninput={handleTitleChange}
    placeholder="Source description"
  />
  {#if origin.url}
    <div class="url-row">
      <a href={origin.url} class="url-link" onclick={handleUrlClick}>
        {origin.url}
      </a>
      <button class="url-clear" onclick={clearUrl} aria-label="Remove URL">&times;</button>
    </div>
  {/if}
</div>

<style>
  .source-editor {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .source-input {
    width: 100%;
    padding: 0.5rem 0.75rem;
    border: 1px solid #ccc;
    border-radius: 6px;
    background: transparent;
    color: inherit;
    font-size: 0.9rem;
    font-family: inherit;
    outline: none;
  }

  .source-editor.dark .source-input {
    border-color: #555;
  }

  .source-input::placeholder {
    color: #888;
  }

  .url-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .url-link {
    flex: 1;
    font-size: 0.85rem;
    color: #0066cc;
    text-decoration: underline;
    word-wrap: break-word;
    overflow-wrap: break-word;
    word-break: break-word;
  }

  .source-editor.dark .url-link {
    color: #6ab0ff;
  }

  .url-clear {
    display: none;
    background: none;
    border: none;
    color: #999;
    font-size: 1.1rem;
    cursor: pointer;
    padding: 0.25rem;
    line-height: 1;
    flex-shrink: 0;
  }

  .url-row:hover .url-clear {
    display: block;
  }

  .url-clear:hover {
    color: #e55;
  }
</style>
