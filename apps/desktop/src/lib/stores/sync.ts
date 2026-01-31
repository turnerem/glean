import { writable } from 'svelte/store';
import { invoke } from '@tauri-apps/api/core';

export interface SyncStatus {
  enabled: boolean;
  authenticated: boolean;
}

export interface SyncState {
  status: SyncStatus;
  isSyncing: boolean;
  lastError: string | null;
  serverUrl: string;
}

function createSyncStore() {
  const initial: SyncState = {
    status: { enabled: false, authenticated: false },
    isSyncing: false,
    lastError: null,
    serverUrl: 'http://localhost:8080',
  };

  const { subscribe, set, update } = writable<SyncState>(initial);

  return {
    subscribe,

    async refreshStatus() {
      try {
        const status = await invoke<SyncStatus>('sync_get_status');
        update(s => ({ ...s, status, lastError: null }));
      } catch (e) {
        update(s => ({ ...s, lastError: String(e) }));
      }
    },

    async configure(serverUrl: string, enabled: boolean) {
      try {
        await invoke('sync_configure', { serverUrl, enabled });
        update(s => ({ ...s, serverUrl, status: { ...s.status, enabled }, lastError: null }));
      } catch (e) {
        update(s => ({ ...s, lastError: String(e) }));
        throw e;
      }
    },

    async login(email: string, password: string) {
      try {
        await invoke('sync_login', { email, password });
        update(s => ({ ...s, status: { ...s.status, authenticated: true }, lastError: null }));
      } catch (e) {
        update(s => ({ ...s, lastError: String(e) }));
        throw e;
      }
    },

    async register(email: string, password: string) {
      try {
        await invoke('sync_register', { email, password });
        update(s => ({ ...s, status: { ...s.status, authenticated: true }, lastError: null }));
      } catch (e) {
        update(s => ({ ...s, lastError: String(e) }));
        throw e;
      }
    },

    async logout() {
      try {
        await invoke('sync_logout');
        update(s => ({ ...s, status: { ...s.status, authenticated: false }, lastError: null }));
      } catch (e) {
        update(s => ({ ...s, lastError: String(e) }));
      }
    },

    async syncNow() {
      update(s => ({ ...s, isSyncing: true, lastError: null }));
      try {
        const result = await invoke('sync_now');
        update(s => ({ ...s, isSyncing: false }));
        return result;
      } catch (e) {
        update(s => ({ ...s, isSyncing: false, lastError: String(e) }));
        throw e;
      }
    },
  };
}

export const syncStore = createSyncStore();
