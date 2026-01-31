<script lang="ts">
  import { tags } from '../stores/notes';
  import type { Tag } from '../types';

  interface Props {
    selected: string[];
    onchange: (tags: string[]) => void;
  }

  let { selected, onchange }: Props = $props();

  let newTagName = $state('');
  let isAddingNew = $state(false);
  let inputRef: HTMLInputElement | null = $state(null);

  function toggleTag(tagName: string) {
    if (selected.includes(tagName)) {
      onchange(selected.filter(t => t !== tagName));
    } else {
      onchange([...selected, tagName]);
    }
  }

  async function addNewTag() {
    if (!newTagName.trim()) return;

    const tag = await tags.create(newTagName.trim());
    if (tag) {
      onchange([...selected, tag.name]);
      newTagName = '';
      isAddingNew = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault();
      addNewTag();
    } else if (e.key === 'Escape') {
      isAddingNew = false;
      newTagName = '';
    }
  }

  function startAddingNew() {
    isAddingNew = true;
    // Focus input after DOM update
    setTimeout(() => inputRef?.focus(), 0);
  }
</script>

<div class="tag-selector">
  <div class="tags-list">
    {#each $tags as tag (tag.id)}
      <button
        class="tag"
        class:selected={selected.includes(tag.name)}
        onclick={() => toggleTag(tag.name)}
        style={tag.color ? `--tag-color: ${tag.color}` : ''}
      >
        {tag.name}
      </button>
    {/each}

    {#if isAddingNew}
      <input
        bind:this={inputRef}
        bind:value={newTagName}
        onkeydown={handleKeydown}
        onblur={() => { if (!newTagName) isAddingNew = false; }}
        placeholder="Tag name..."
        class="new-tag-input"
      />
    {:else}
      <button class="add-tag" onclick={startAddingNew}>
        + New
      </button>
    {/if}
  </div>
</div>

<style>
  .tag-selector {
    margin: 0.5rem 0;
  }

  .tags-list {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .tag {
    padding: 0.25rem 0.75rem;
    border-radius: 1rem;
    border: 1px solid #444;
    background: transparent;
    color: #ccc;
    font-size: 0.8rem;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .tag:hover {
    border-color: #666;
    color: #fff;
  }

  .tag.selected {
    background: var(--tag-color, #4a5568);
    border-color: var(--tag-color, #4a5568);
    color: #fff;
  }

  .add-tag {
    padding: 0.25rem 0.75rem;
    border-radius: 1rem;
    border: 1px dashed #444;
    background: transparent;
    color: #888;
    font-size: 0.8rem;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .add-tag:hover {
    border-color: #666;
    color: #ccc;
  }

  .new-tag-input {
    padding: 0.25rem 0.75rem;
    border-radius: 1rem;
    border: 1px solid #4a9eff;
    background: #1a1a2e;
    color: #fff;
    font-size: 0.8rem;
    width: 100px;
    outline: none;
  }
</style>
