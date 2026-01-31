<script lang="ts">
  import type { Origin } from '../types';

  interface Props {
    origin: Origin;
    onchange?: (origin: Origin) => void;
    editable?: boolean;
  }

  let { origin, onchange, editable = false }: Props = $props();

  let isEditing = $state(false);
  let manualInput = $state('');

  function getDisplayText(): string {
    switch (origin.type) {
      case 'url':
        return origin.title || origin.url || 'Web page';
      case 'book':
        return origin.book_title || 'Book';
      case 'manual':
        return origin.raw_input || 'Manual entry';
      default:
        return 'Unknown source';
    }
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

  function startEditing() {
    if (!editable) return;
    isEditing = true;
    manualInput = origin.raw_input || '';
  }

  function saveManualOrigin() {
    if (onchange && manualInput.trim()) {
      onchange({
        type: 'manual',
        raw_input: manualInput.trim(),
      });
    }
    isEditing = false;
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault();
      saveManualOrigin();
    } else if (e.key === 'Escape') {
      isEditing = false;
    }
  }
</script>

<div class="origin-display">
  {#if isEditing}
    <input
      bind:value={manualInput}
      onkeydown={handleKeydown}
      onblur={saveManualOrigin}
      placeholder="Enter source (URL, book title, etc.)"
      class="origin-input"
      autofocus
    />
  {:else}
    <button
      class="origin-badge"
      class:editable
      onclick={startEditing}
      disabled={!editable}
    >
      <span class="icon">{getIcon()}</span>
      <span class="text">{getDisplayText()}</span>
      {#if editable && origin.type === 'unknown'}
        <span class="edit-hint">click to edit</span>
      {/if}
    </button>
  {/if}
</div>

<style>
  .origin-display {
    margin: 0.5rem 0;
  }

  .origin-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.375rem 0.75rem;
    border-radius: 0.5rem;
    border: 1px solid #333;
    background: #252538;
    color: #aaa;
    font-size: 0.8rem;
    cursor: default;
    max-width: 100%;
    overflow: hidden;
  }

  .origin-badge.editable {
    cursor: pointer;
  }

  .origin-badge.editable:hover {
    border-color: #444;
    color: #ccc;
  }

  .icon {
    flex-shrink: 0;
  }

  .text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .edit-hint {
    color: #666;
    font-size: 0.7rem;
    font-style: italic;
  }

  .origin-input {
    width: 100%;
    padding: 0.5rem;
    border-radius: 0.5rem;
    border: 1px solid #4a9eff;
    background: #1a1a2e;
    color: #fff;
    font-size: 0.875rem;
    outline: none;
  }

  .origin-input::placeholder {
    color: #666;
  }
</style>
