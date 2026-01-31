// Shared types matching Rust models

export interface Origin {
  type: 'url' | 'book' | 'manual' | 'unknown';
  url?: string;
  title?: string;
  book_title?: string;
  chapter?: string;
  page?: string;
  raw_input?: string;
}

export interface Note {
  id: string;
  content: string;
  image_data?: string; // Base64 encoded image
  tags: string[];
  origin: Origin;
  created_at: string;
  updated_at: string;
  sync_version: number;
  is_deleted: boolean;
}

export interface Tag {
  id: string;
  name: string;
  color?: string;
  created_at: string;
}

export interface CreateNoteRequest {
  content: string;
  image_data?: string;
  tags: string[];
  origin: Origin;
}

export interface CreateTagRequest {
  name: string;
  color?: string;
}
