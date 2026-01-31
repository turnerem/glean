<script lang="ts">
  import { capture, notes, tags } from '../stores/notes';
  import TagSelector from './TagSelector.svelte';
  import OriginDisplay from './OriginDisplay.svelte';
  import type { Origin } from '../types';

  let isSaving = $state(false);
  let textareaRef: HTMLTextAreaElement | null = $state(null);

  // Focus textarea when popup opens
  $effect(() => {
    if ($capture.isOpen && textareaRef) {
      textareaRef.focus();
      textareaRef.select();
    }
  });

  async function save() {
    if (!$capture.content.trim() || isSaving) return;

    isSaving = true;
    try {
      const note = await notes.create(
        $capture.content,
        $capture.selectedTags,
        $capture.origin
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

  function handleKeydown(e: KeyboardEvent) {
    // Cmd/Ctrl+Enter to save
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      e.preventDefault();
      save();
    } else if (e.key === 'Escape') {
      e.preventDefault();
      cancel();
    }
  }

  function handleTagsChange(newTags: string[]) {
    capture.setTags(newTags);
  }

  function handleOriginChange(newOrigin: Origin) {
    capture.setOrigin(newOrigin);
  }

  function handleContentChange(e: Event) {
    const target = e.target as HTMLTextAreaElement;
    capture.setContent(target.value);
  }
</script>

{#if $capture.isOpen}
  <div class="overlay" onclick={cancel} role="presentation">
    <div class="popup" onclick={(e) => e.stopPropagation()} role="dialog" aria-modal="true">
      <div class="header">
        <h2>Save Note</h2>
        <button class="close-btn" onclick={cancel} aria-label="Close">×</button>
      </div>

      <div class="content">
        <textarea
          bind:this={textareaRef}
          value={$capture.content}
          oninput={handleContentChange}
          onkeydown={handleKeydown}
          placeholder="Captured text..."
          rows="4"
        ></textarea>

        <div class="section">
          <label>Source</label>
          <OriginDisplay
            origin={$capture.origin}
            onchange={handleOriginChange}
            editable={true}
          />
        </div>

        <div class="section">
          <label>Tags</label>
          <TagSelector
            selected={$capture.selectedTags}
            onchange={handleTagsChange}
          />
        </div>
      </div>

      <div class="footer">
        <span class="hint">⌘+Enter to save, Esc to cancel</span>
        <div class="actions">
          <button class="btn-secondary" onclick={cancel}>Cancel</button>
          <button
            class="btn-primary"
            onclick={save}
            disabled={!$capture.content.trim() || isSaving}
          >
            {isSaving ? 'Saving...' : 'Save'}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .popup {
    background: #1e1e32;
    border-radius: 12px;
    width: 90%;
    max-width: 400px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
    border: 1px solid #333;
    overflow: hidden;
  }

  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1rem;
    border-bottom: 1px solid #333;
  }

  .header h2 {
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    color: #fff;
  }

  .close-btn {
    background: none;
    border: none;
    color: #666;
    font-size: 1.5rem;
    cursor: pointer;
    padding: 0;
    line-height: 1;
  }

  .close-btn:hover {
    color: #fff;
  }

  .content {
    padding: 1rem;
  }

  textarea {
    width: 100%;
    padding: 0.75rem;
    border-radius: 8px;
    border: 1px solid #333;
    background: #252538;
    color: #fff;
    font-size: 0.9rem;
    font-family: inherit;
    resize: vertical;
    min-height: 80px;
  }

  textarea:focus {
    outline: none;
    border-color: #4a9eff;
  }

  textarea::placeholder {
    color: #666;
  }

  .section {
    margin-top: 1rem;
  }

  .section label {
    display: block;
    font-size: 0.75rem;
    color: #888;
    margin-bottom: 0.375rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1rem;
    border-top: 1px solid #333;
    background: #1a1a2e;
  }

  .hint {
    font-size: 0.75rem;
    color: #666;
  }

  .actions {
    display: flex;
    gap: 0.5rem;
  }

  .btn-secondary,
  .btn-primary {
    padding: 0.5rem 1rem;
    border-radius: 6px;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .btn-secondary {
    background: transparent;
    border: 1px solid #444;
    color: #ccc;
  }

  .btn-secondary:hover {
    border-color: #666;
    color: #fff;
  }

  .btn-primary {
    background: #4a9eff;
    border: none;
    color: #fff;
  }

  .btn-primary:hover:not(:disabled) {
    background: #3a8eef;
  }

  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
