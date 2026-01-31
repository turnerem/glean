<script lang="ts">
  import { onMount } from 'svelte';
  import { capture, notes, tags } from '../stores/notes';
  import { theme } from '../stores/theme';
  import SourceEditor from './SourceEditor.svelte';
  import type { Origin } from '../types';

  let isSaving = $state(false);
  let textareaRef: HTMLTextAreaElement | null = $state(null);
  let hasInitialFocus = $state(false);
  let newTagName = $state('');
  let isAddingTag = $state(false);
  let tagInputRef: HTMLInputElement | null = $state(null);

  // Focus textarea only once when popup opens
  $effect(() => {
    if ($capture.isOpen && textareaRef && !hasInitialFocus) {
      hasInitialFocus = true;
      textareaRef.focus();
      textareaRef.selectionStart = textareaRef.value.length;
      textareaRef.selectionEnd = textareaRef.value.length;
    }
    if (!$capture.isOpen) {
      hasInitialFocus = false;
    }
  });

  // Global keyboard handler
  onMount(() => {
    function handleGlobalKeydown(e: KeyboardEvent) {
      if (!$capture.isOpen) return;

      if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        save();
      } else if (e.key === 'Escape') {
        e.preventDefault();
        cancel();
      }
    }

    window.addEventListener('keydown', handleGlobalKeydown);
    return () => window.removeEventListener('keydown', handleGlobalKeydown);
  });

  async function save() {
    if (isSaving) return;
    if (!$capture.content.trim() && !$capture.imageData) return;

    isSaving = true;
    try {
      const note = await notes.create(
        $capture.content,
        $capture.selectedTags,
        $capture.origin,
        $capture.imageData
      );

      if (note) {
        capture.close();
      }
    } finally {
      isSaving = false;
    }
  }

  function cancel() {
    capture.close();
  }

  function handleContentChange(e: Event) {
    const target = e.target as HTMLTextAreaElement;
    capture.setContent(target.value);
  }

  function handleOriginChange(origin: Origin) {
    capture.setOrigin(origin);
  }

  function removeTag(tagName: string) {
    capture.setTags($capture.selectedTags.filter(t => t !== tagName));
  }

  function getTagColorByName(tagName: string): string {
    const tagIndex = $tags.findIndex(t => t.name === tagName);
    return tagColors[tagIndex >= 0 ? tagIndex % tagColors.length : 0];
  }

  // Determine if image is portrait or landscape
  function isPortrait(dataUrl: string): Promise<boolean> {
    return new Promise((resolve) => {
      const img = new Image();
      img.onload = () => resolve(img.height > img.width);
      img.src = dataUrl;
    });
  }

  let imageOrientation = $state<'portrait' | 'landscape'>('landscape');

  $effect(() => {
    if ($capture.imageData) {
      isPortrait($capture.imageData).then(portrait => {
        imageOrientation = portrait ? 'portrait' : 'landscape';
      });
    }
  });

  // Tag colors for display
  const tagColors = ['#fff3b0', '#ffc0cb', '#b0e0e6', '#98fb98', '#dda0dd'];

  function startAddingTag() {
    isAddingTag = true;
    setTimeout(() => tagInputRef?.focus(), 0);
  }

  async function createNewTag() {
    if (!newTagName.trim()) return;
    const tag = await tags.create(newTagName.trim());
    if (tag) {
      capture.setTags([...$capture.selectedTags, tag.name]);
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
      t.name.toLowerCase().includes(search) && !$capture.selectedTags.includes(t.name)
    );
  });

  function selectSuggestion(tagName: string) {
    capture.setTags([...$capture.selectedTags, tagName]);
    newTagName = '';
    isAddingTag = false;
  }
</script>

{#if $capture.isOpen}
  <div class="overlay" class:dark={$theme === 'dark'} onclick={cancel} role="presentation">
    <div class="popup" class:dark={$theme === 'dark'} onclick={(e) => e.stopPropagation()} role="dialog" aria-modal="true">

      {#if $capture.imageData}
        <div class="content-area" class:portrait={imageOrientation === 'portrait'}>
          <div class="image-container">
            <img src={$capture.imageData} alt="Captured" class="captured-image" />
          </div>
          <div class="text-beside">
            <textarea
              bind:this={textareaRef}
              value={$capture.content}
              oninput={handleContentChange}
              placeholder="Add a description..."
              class="note-textarea"
              rows="4"
            ></textarea>
          </div>
        </div>
      {:else}
        <div class="content-area text-only">
          <textarea
            bind:this={textareaRef}
            value={$capture.content}
            oninput={handleContentChange}
            placeholder="Note text..."
            class="note-textarea full"
            rows="6"
          ></textarea>
        </div>
      {/if}

      <div class="meta-section">
        <SourceEditor origin={$capture.origin} onOriginChange={handleOriginChange} />

        <div class="tags-row">
          {#each $capture.selectedTags as tagName}
            <span class="tag-chip" style="background: {getTagColorByName(tagName)}">
              {tagName}
              <button class="tag-remove" onclick={() => removeTag(tagName)}>&times;</button>
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
      </div>

      <div class="footer">
        <span class="hint">Cmd+Return to save</span>
      </div>
    </div>
  </div>
{/if}

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.3);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .popup {
    background: #e8e8e8;
    border-radius: 8px;
    width: 90%;
    max-width: 500px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
    overflow: hidden;
    color: #222;
  }

  .popup.dark {
    background: #2a2a3e;
    color: #eee;
  }

  .content-area {
    padding: 1.5rem;
  }

  .content-area.portrait {
    display: flex;
    gap: 1rem;
  }

  .content-area.text-only {
    padding: 1.5rem;
  }

  .image-container {
    flex-shrink: 0;
  }

  .content-area.portrait .image-container {
    width: 45%;
  }

  .content-area:not(.portrait) .image-container {
    margin-bottom: 1rem;
  }

  .captured-image {
    width: 100%;
    height: auto;
    border-radius: 4px;
    border: 2px solid #ccc;
    display: block;
  }

  .popup.dark .captured-image {
    border-color: #555;
  }

  .text-beside {
    flex: 1;
    display: flex;
    flex-direction: column;
  }

  .note-textarea {
    width: 100%;
    padding: 0;
    border: none;
    background: transparent;
    color: inherit;
    font-size: 1rem;
    font-family: inherit;
    resize: none;
    outline: none;
    line-height: 1.5;
  }

  .note-textarea::placeholder {
    color: #888;
  }

  .note-textarea.full {
    min-height: 120px;
  }

  .meta-section {
    padding: 0 1.5rem 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .tags-row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    align-items: center;
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

  .tag-suggestions {
    position: absolute;
    top: 100%;
    left: 0;
    min-width: 120px;
    background: #fff;
    border: 1px solid #ccc;
    border-radius: 4px;
    margin-top: 2px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.15);
    z-index: 10;
    max-height: 150px;
    overflow-y: auto;
  }

  .popup.dark .tag-suggestions {
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

  .popup.dark .tag-suggestion:hover {
    background: #3a3a4e;
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

  .popup.dark .add-tag-btn {
    border-color: #555;
    color: #888;
  }

  .popup.dark .add-tag-btn:hover {
    border-color: #777;
    color: #ccc;
  }

  .new-tag-input {
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    border: 1px solid #0066cc;
    background: #fff;
    color: #333;
    font-size: 0.85rem;
    width: 100px;
    outline: none;
  }

  .popup.dark .new-tag-input {
    background: #1a1a2e;
    color: #eee;
    border-color: #6ab0ff;
  }

  .footer {
    padding: 0.75rem 1.5rem;
    border-top: 1px solid #ccc;
    text-align: center;
  }

  .popup.dark .footer {
    border-color: #444;
  }

  .hint {
    font-size: 0.75rem;
    color: #888;
  }
</style>
