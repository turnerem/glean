import { writable } from 'svelte/store';
import { invoke } from '@tauri-apps/api/core';
import type { Note, Tag, Origin, CreateNoteRequest, CreateTagRequest } from '../types';

// Notes store
function createNotesStore() {
  const { subscribe, set, update } = writable<Note[]>([]);

  return {
    subscribe,
    async load() {
      try {
        const notes = await invoke<Note[]>('get_notes');
        set(notes);
      } catch (e) {
        console.error('Failed to load notes:', e);
      }
    },
    async create(content: string, tags: string[], origin: Origin): Promise<Note | null> {
      try {
        const request: CreateNoteRequest = { content, tags, origin };
        const note = await invoke<Note>('create_note', { request });
        update(notes => [note, ...notes]);
        return note;
      } catch (e) {
        console.error('Failed to create note:', e);
        return null;
      }
    },
    async delete(id: string) {
      try {
        await invoke('delete_note', { id });
        update(notes => notes.filter(n => n.id !== id));
      } catch (e) {
        console.error('Failed to delete note:', e);
      }
    },
  };
}

// Tags store
function createTagsStore() {
  const { subscribe, set, update } = writable<Tag[]>([]);

  return {
    subscribe,
    async load() {
      try {
        const tags = await invoke<Tag[]>('get_tags');
        set(tags);
      } catch (e) {
        console.error('Failed to load tags:', e);
      }
    },
    async create(name: string, color?: string): Promise<Tag | null> {
      try {
        const request: CreateTagRequest = { name, color };
        const tag = await invoke<Tag>('create_tag', { request });
        update(tags => [...tags, tag]);
        return tag;
      } catch (e) {
        console.error('Failed to create tag:', e);
        return null;
      }
    },
  };
}

// Current capture state
interface CaptureState {
  isOpen: boolean;
  content: string;
  origin: Origin;
  selectedTags: string[];
}

function createCaptureStore() {
  const initial: CaptureState = {
    isOpen: false,
    content: '',
    origin: { type: 'unknown' },
    selectedTags: [],
  };

  const { subscribe, set, update } = writable<CaptureState>(initial);

  return {
    subscribe,
    open(content: string, origin: Origin) {
      set({
        isOpen: true,
        content,
        origin,
        selectedTags: [],
      });
    },
    close() {
      set(initial);
    },
    setTags(tags: string[]) {
      update(s => ({ ...s, selectedTags: tags }));
    },
    setContent(content: string) {
      update(s => ({ ...s, content }));
    },
    setOrigin(origin: Origin) {
      update(s => ({ ...s, origin }));
    },
  };
}

export const notes = createNotesStore();
export const tags = createTagsStore();
export const capture = createCaptureStore();
