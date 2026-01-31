import { writable } from 'svelte/store';

export type Theme = 'light' | 'dark';

function createThemeStore() {
  // Check localStorage for saved preference, default to light
  const stored = typeof localStorage !== 'undefined' ? localStorage.getItem('theme') : null;
  const initial: Theme = (stored === 'dark' || stored === 'light') ? stored : 'light';

  const { subscribe, set } = writable<Theme>(initial);

  return {
    subscribe,
    set(theme: Theme) {
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('theme', theme);
      }
      set(theme);
    },
    toggle() {
      let current: Theme = 'light';
      subscribe(v => current = v)();
      const next: Theme = current === 'light' ? 'dark' : 'light';
      this.set(next);
    }
  };
}

export const theme = createThemeStore();
