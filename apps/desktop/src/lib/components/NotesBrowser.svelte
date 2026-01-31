<script lang="ts">
  import { openUrl as tauriOpenUrl } from '@tauri-apps/plugin-opener';
  import { notes, tags } from '../stores/notes';
  import { theme } from '../stores/theme';
  import SourceEditor from './SourceEditor.svelte';
  import type { Note, Origin } from '../types';

  interface Props {
    onBack: () => void;
  }

  let { onBack }: Props = $props();

  let sortOrder = $state<'newest' | 'oldest'>('newest');
  let selectedTagFilter = $state<string | null>(null);
  let editingNoteId = $state<string | null>(null);
  let editContent = $state('');
  let editTags = $state<string[]>([]);
  let editOrigin = $state<Origin>({ type: 'unknown' });
  let isSaving = $state(false);
  let isAddingTag = $state(false);
  let newTagName = $state('');
  let tagInputRef: HTMLInputElement | null = $state(null);

  let filteredNotes = $derived.by(() => {
    let result = [...$notes];

    if (selectedTagFilter) {
      result = result.filter(n => n.tags.includes(selectedTagFilter!));
    }

    result.sort((a, b) => {
      const dateA = new Date(a.created_at).getTime();
      const dateB = new Date(b.created_at).getTime();
      return sortOrder === 'newest' ? dateB - dateA : dateA - dateB;
    });

    return result;
  });

  function formatDate(dateStr: string): string {
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  }

  function startEditing(note: Note) {
    editingNoteId = note.id;
    editContent = note.content;
    editTags = [...note.tags];
    editOrigin = { ...note.origin };
  }

  function cancelEditing() {
    editingNoteId = null;
    editContent = '';
    editTags = [];
    editOrigin = { type: 'unknown' };
  }

  async function saveEdit() {
    if (!editingNoteId || isSaving) return;

    isSaving = true;
    const success = await notes.updateNote(editingNoteId, editContent, editTags, editOrigin);
    isSaving = false;

    if (success) {
      cancelEditing();
    }
  }

  function handleOriginChange(origin: Origin) {
    editOrigin = origin;
  }

  async function deleteNote(e: MouseEvent, id: string) {
    e.stopPropagation();
    await notes.delete(id);
  }

  async function openUrl(e: MouseEvent, url: string) {
    e.preventDefault();
    e.stopPropagation();
    try {
      await tauriOpenUrl(url);
    } catch (err) {
      console.error('Failed to open URL:', err);
      window.open(url, '_blank');
    }
  }

  function getSourceTitle(origin: Origin): string {
    if (origin.title) return origin.title;
    if (origin.url) {
      try {
        return new URL(origin.url).hostname;
      } catch {
        return origin.url;
      }
    }
    return '';
  }

  const tagColors = ['#fff3b0', '#ffc0cb', '#b0e0e6', '#98fb98', '#dda0dd'];
  function getTagColor(tagName: string): string {
    const tagIndex = $tags.findIndex(t => t.name === tagName);
    return tagColors[tagIndex >= 0 ? tagIndex % tagColors.length : 0];
  }

  function startAddingTag() {
    isAddingTag = true;
    setTimeout(() => tagInputRef?.focus(), 0);
  }

  async function createNewTag() {
    if (!newTagName.trim()) return;
    const tag = await tags.create(newTagName.trim());
    if (tag) {
      editTags = [...editTags, tag.name];
      newTagName = '';
      isAddingTag = false;
    }
  }

  function handleTagInputKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault();
      e.stopPropagation();
      if (filteredSuggestions.length > 0) {
        selectSuggestion(filteredSuggestions[0].name);
      } else {
        createNewTag();
      }
    } else if (e.key === 'Escape') {
      isAddingTag = false;
      newTagName = '';
    }
  }

  let filteredSuggestions = $derived.by(() => {
    if (!newTagName.trim()) return [];
    const search = newTagName.toLowerCase();
    return $tags.filter(t =>
      t.name.toLowerCase().includes(search) && !editTags.includes(t.name)
    );
  });

  function selectSuggestion(tagName: string) {
    editTags = [...editTags, tagName];
    newTagName = '';
    isAddingTag = false;
  }

  function removeEditTag(tagName: string) {
    editTags = editTags.filter(t => t !== tagName);
  }
</script>

<div class="browser" class:dark={$theme === 'dark'}>
  <header class="browser-header">
    <button class="back-btn" onclick={onBack}>Back</button>
    <h1>All Notes</h1>
    <select bind:value={sortOrder} class="sort-select">
      <option value="newest">Newest</option>
      <option value="oldest">Oldest</option>
    </select>
  </header>

  <div class="tag-filter">
    <button
      class="filter-btn"
      class:selected={selectedTagFilter === null}
      onclick={() => selectedTagFilter = null}
    >
      All
    </button>
    {#each $tags as tag, i (tag.id)}
      <button
        class="filter-btn"
        class:selected={selectedTagFilter === tag.name}
        style="--tag-bg: {tagColors[i % tagColors.length]}"
        onclick={() => selectedTagFilter = tag.name}
      >
        {tag.name}
      </button>
    {/each}
  </div>

  <div class="notes-list">
    {#if filteredNotes.length === 0}
      <div class="empty">
        <p>No notes found</p>
      </div>
    {:else}
      {#each filteredNotes as note (note.id)}
        <div class="note-card" class:editing={editingNoteId === note.id}>
          {#if editingNoteId === note.id}
            <div class="edit-mode">
              {#if note.image_data}
                <img src={note.image_data} alt="Note" class="note-image" />
              {/if}
              <textarea
                bind:value={editContent}
                class="edit-textarea"
                rows="6"
                placeholder="Note text..."
              ></textarea>

              <SourceEditor origin={editOrigin} onOriginChange={handleOriginChange} />

              <div class="edit-tags">
                {#each editTags as tagName}
                  <span class="tag-chip" style="background: {getTagColor(tagName)}">
                    {tagName}
                    <button class="tag-remove" onclick={() => removeEditTag(tagName)}>&times;</button>
                  </span>
                {/each}
                {#if isAddingTag}
                  <div class="tag-input-wrapper">
                    <input
                      bind:this={tagInputRef}
                      bind:value={newTagName}
                      onkeydown={handleTagInputKeydown}
                      onblur={() => { setTimeout(() => { if (!newTagName) isAddingTag = false; }, 150); }}
                      placeholder="Tag name..."
                      class="new-tag-input"
                    />
                    {#if filteredSuggestions.length > 0}
                      <div class="tag-suggestions">
                        {#each filteredSuggestions as suggestion}
                          <button
                            class="tag-suggestion"
                            onmousedown={() => selectSuggestion(suggestion.name)}
                          >
                            {suggestion.name}
                          </button>
                        {/each}
                      </div>
                    {/if}
                  </div>
                {:else}
                  <button class="add-tag-btn" onclick={startAddingTag}>+ Add</button>
                {/if}
              </div>

              <div class="edit-actions">
                <button class="btn cancel" onclick={cancelEditing}>Cancel</button>
                <button class="btn save" onclick={saveEdit} disabled={isSaving}>
                  {isSaving ? 'Saving...' : 'Save'}
                </button>
              </div>
            </div>
          {:else}
            <div class="note-view" onclick={() => startEditing(note)}>
              {#if note.image_data}
                <img src={note.image_data} alt="Note" class="note-image" />
              {/if}

              {#if note.content}
                <p class="note-text">{note.content}</p>
              {/if}

              {#if getSourceTitle(note.origin)}
                <div class="source-section">
                  {#if note.origin.url}
                    <a href={note.origin.url} class="source-link" onclick={(e) => openUrl(e, note.origin.url!)}>
                      {getSourceTitle(note.origin)}
                    </a>
                  {:else}
                    <span class="source-text">{getSourceTitle(note.origin)}</span>
                  {/if}
                </div>
              {/if}

              {#if note.tags.length > 0}
                <div class="note-tags">
                  {#each note.tags as tag}
                    <span class="tag-display" style="background: {getTagColor(tag)}">{tag}</span>
                  {/each}
                </div>
              {/if}

              <div class="note-footer">
                <span class="date">{formatDate(note.created_at)}</span>
                <button class="delete-btn" onclick={(e) => deleteNote(e, note.id)}>Delete</button>
              </div>
            </div>
          {/if}
        </div>
      {/each}
    {/if}
  </div>
</div>

<style>
  .browser {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    color: #222;
  }

  .browser.dark {
    color: #eee;
  }

  .browser-header {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .browser-header h1 {
    flex: 1;
    margin: 0;
    font-size: 1.25rem;
    font-weight: 600;
  }

  .back-btn {
    padding: 0.5rem 1rem;
    background: transparent;
    border: 1px solid #aaa;
    border-radius: 6px;
    color: inherit;
    cursor: pointer;
    font-size: 0.9rem;
  }

  .browser.dark .back-btn {
    border-color: #555;
  }

  .sort-select {
    padding: 0.5rem 0.75rem;
    background: transparent;
    border: 1px solid #aaa;
    border-radius: 6px;
    color: inherit;
    font-size: 0.9rem;
  }

  .browser.dark .sort-select {
    border-color: #555;
    background: #333;
  }

  .tag-filter {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .filter-btn {
    padding: 0.35rem 0.85rem;
    border-radius: 6px;
    border: none;
    background: #333;
    color: #fff;
    font-size: 0.85rem;
    cursor: pointer;
    opacity: 0.4;
  }

  .filter-btn:hover {
    opacity: 0.7;
  }

  .filter-btn.selected {
    opacity: 1;
    background: var(--tag-bg, #333);
    color: #333;
  }

  .notes-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .note-card {
    background: #e8e8e8;
    border-radius: 10px;
    padding: 1.25rem;
    cursor: pointer;
    overflow: hidden;
  }

  .browser.dark .note-card {
    background: #2a2a3e;
  }

  .note-card.editing {
    cursor: default;
  }

  .note-image {
    width: 100%;
    max-height: 250px;
    object-fit: contain;
    border-radius: 6px;
    margin-bottom: 1rem;
  }

  .note-text {
    margin: 0 0 1rem 0;
    font-size: 1.05rem;
    line-height: 1.6;
    white-space: pre-wrap;
    word-wrap: break-word;
    overflow-wrap: break-word;
    word-break: break-word;
  }

  .source-section {
    margin-bottom: 0.75rem;
    max-width: 100%;
    overflow: hidden;
  }

  .source-text {
    display: block;
    font-size: 0.9rem;
    color: #777;
    word-wrap: break-word;
    overflow-wrap: break-word;
    word-break: break-word;
  }

  .browser.dark .source-text {
    color: #999;
  }

  .source-link {
    display: block;
    font-size: 0.9rem;
    color: #0066cc;
    text-decoration: underline;
    word-wrap: break-word;
    overflow-wrap: break-word;
    word-break: break-word;
    cursor: pointer;
  }

  .source-link:hover {
    color: #0052a3;
  }

  .browser.dark .source-link {
    color: #6ab0ff;
  }

  .browser.dark .source-link:hover {
    color: #8ec4ff;
  }

  .note-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
  }

  .tag-display {
    padding: 0.25rem 0.75rem;
    border-radius: 4px;
    font-size: 0.85rem;
    color: #333;
  }

  .tag-pill {
    padding: 0.25rem 0.75rem;
    border-radius: 4px;
    border: none;
    background: var(--tag-bg);
    color: #333;
    font-size: 0.85rem;
    cursor: pointer;
    opacity: 0.4;
  }

  .tag-pill:hover {
    opacity: 0.7;
  }

  .tag-pill.selected {
    opacity: 1;
  }

  .note-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .date {
    font-size: 0.85rem;
    color: #888;
  }

  .delete-btn {
    padding: 0.35rem 0.75rem;
    background: transparent;
    border: 1px solid #bbb;
    border-radius: 6px;
    color: #888;
    font-size: 0.85rem;
    cursor: pointer;
  }

  .delete-btn:hover {
    border-color: #e55;
    color: #e55;
  }

  .edit-mode {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .edit-textarea {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #ccc;
    border-radius: 6px;
    background: transparent;
    color: inherit;
    font-size: 1rem;
    font-family: inherit;
    resize: vertical;
    min-height: 120px;
  }

  .browser.dark .edit-textarea {
    border-color: #555;
  }

  .edit-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .add-tag-btn {
    padding: 0.25rem 0.75rem;
    border-radius: 4px;
    border: 1px dashed #999;
    background: transparent;
    color: #666;
    font-size: 0.85rem;
    cursor: pointer;
  }

  .add-tag-btn:hover {
    border-color: #666;
    color: #333;
  }

  .browser.dark .add-tag-btn {
    border-color: #555;
    color: #888;
  }

  .browser.dark .add-tag-btn:hover {
    border-color: #777;
    color: #ccc;
  }

  .tag-chip {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.85rem;
    color: #333;
  }

  .tag-remove {
    background: none;
    border: none;
    color: #666;
    font-size: 1rem;
    cursor: pointer;
    padding: 0;
    line-height: 1;
  }

  .tag-remove:hover {
    color: #333;
  }

  .tag-input-wrapper {
    position: relative;
  }

  .new-tag-input {
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    border: 1px solid #0066cc;
    background: #fff;
    color: #333;
    font-size: 0.85rem;
    width: 120px;
    outline: none;
  }

  .browser.dark .new-tag-input {
    background: #1a1a2e;
    color: #eee;
    border-color: #6ab0ff;
  }

  .tag-suggestions {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    background: #fff;
    border: 1px solid #ccc;
    border-radius: 4px;
    margin-top: 2px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.15);
    z-index: 10;
    max-height: 150px;
    overflow-y: auto;
  }

  .browser.dark .tag-suggestions {
    background: #2a2a3e;
    border-color: #555;
  }

  .tag-suggestion {
    display: block;
    width: 100%;
    padding: 0.5rem;
    border: none;
    background: none;
    color: inherit;
    font-size: 0.85rem;
    text-align: left;
    cursor: pointer;
  }

  .tag-suggestion:hover {
    background: #f0f0f0;
  }

  .browser.dark .tag-suggestion:hover {
    background: #3a3a4e;
  }

  .edit-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    margin-top: 0.5rem;
  }

  .btn {
    padding: 0.5rem 1rem;
    border-radius: 6px;
    font-size: 0.9rem;
    cursor: pointer;
  }

  .btn.cancel {
    background: transparent;
    border: 1px solid #aaa;
    color: inherit;
  }

  .btn.save {
    background: #333;
    border: none;
    color: #fff;
  }

  .browser.dark .btn.save {
    background: #555;
  }

  .btn.save:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .empty {
    text-align: center;
    padding: 3rem;
    color: #888;
  }
</style>
