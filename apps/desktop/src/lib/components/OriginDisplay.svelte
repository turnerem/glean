<script lang="ts">
  import type { Origin } from '../types';

  interface Props {
    origin: Origin;
    onchange?: (origin: Origin) => void;
    editable?: boolean;
  }

  let { origin, onchange, editable = false }: Props = $props();

  let isEditingTitle = $state(false);
  let editTitle = $state('');

  function getDisplayTitle(): string {
    if (origin.title) return origin.title;
    if (origin.url) {
      try {
        return new URL(origin.url).hostname;
      } catch {
        return origin.url;
      }
    }
    if (origin.book_title) return origin.book_title;
    if (origin.raw_input) return origin.raw_input;
    return 'Unknown source';
  }

  function getIcon(): string {
    switch (origin.type) {
      case 'url':
        return '🌐';
      case 'book':
        return '📖';
      case 'manual':
        return '✏️';
      default:
        return '❓';
    }
  }

  function startEditingTitle() {
    if (!editable) return;
    isEditingTitle = true;
    editTitle = origin.title || '';
  }

  function saveTitleEdit() {
    if (!onchange) {
      isEditingTitle = false;
      return;
    }

    const newTitle = editTitle.trim();

    // Preserve existing origin data but update the title
    onchange({
      ...origin,
      title: newTitle || undefined,
    });

    isEditingTitle = false;
  }

  function handleTitleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault();
      saveTitleEdit();
    } else if (e.key === 'Escape') {
      isEditingTitle = false;
    }
  }

  function openUrl(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();
    if (origin.url) {
      window.open(origin.url, '_blank');
    }
  }
</script>

<div class="origin-display">
  <span class="icon">{getIcon()}</span>

  <div class="origin-content">
    {#if isEditingTitle}
      <input
        bind:value={editTitle}
        onkeydown={handleTitleKeydown}
        onblur={saveTitleEdit}
        placeholder="Enter description"
        class="title-input"
        autofocus
      />
    {:else}
      <button
        class="title"
        class:editable
        onclick={startEditingTitle}
        title={editable ? 'Click to edit description' : ''}
      >
        {getDisplayTitle()}
        {#if editable}
          <span class="edit-icon">✎</span>
        {/if}
      </button>
    {/if}

    {#if origin.url}
      <a
        href={origin.url}
        class="url-link"
        onclick={openUrl}
        title={origin.url}
      >
        {#if origin.url.length > 50}
          {origin.url.slice(0, 50)}...
        {:else}
          {origin.url}
        {/if}
      </a>
    {/if}

    {#if origin.type === 'book' && origin.chapter}
      <span class="book-detail">Ch. {origin.chapter}</span>
    {/if}

    {#if origin.type === 'book' && origin.page}
      <span class="book-detail">p. {origin.page}</span>
    {/if}
  </div>
</div>

<style>
  .origin-display {
    display: flex;
    align-items: flex-start;
    gap: 0.5rem;
    padding: 0.5rem 0.75rem;
    border-radius: 0.5rem;
    border: 1px solid #333;
    background: #252538;
  }

  .icon {
    flex-shrink: 0;
    font-size: 1rem;
    line-height: 1.4;
  }

  .origin-content {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    min-width: 0;
    flex: 1;
  }

  .title {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    background: none;
    border: none;
    padding: 0;
    margin: 0;
    color: #ddd;
    font-size: 0.875rem;
    font-weight: 500;
    text-align: left;
    cursor: default;
    word-break: break-word;
  }

  .title.editable {
    cursor: pointer;
  }

  .title.editable:hover {
    color: #fff;
  }

  .title.editable:hover .edit-icon {
    opacity: 1;
  }

  .edit-icon {
    opacity: 0;
    font-size: 0.75rem;
    color: #888;
    transition: opacity 0.15s;
  }

  .url-link {
    color: #4a9eff;
    font-size: 0.75rem;
    text-decoration: none;
    word-break: break-all;
    cursor: pointer;
  }

  .url-link:hover {
    text-decoration: underline;
    color: #6ab0ff;
  }

  .book-detail {
    color: #888;
    font-size: 0.75rem;
  }

  .title-input {
    width: 100%;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    border: 1px solid #4a9eff;
    background: #1a1a2e;
    color: #fff;
    font-size: 0.875rem;
    outline: none;
  }

  .title-input::placeholder {
    color: #666;
  }
</style>
