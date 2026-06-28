import { browser } from "$app/environment";

const MAX_HISTORY = 10;
const STORAGE_PREFIX = "nd.admin.search.";

function loadHistory(mode: string): string[] {
  if (!browser) return [];
  try {
    const raw = localStorage.getItem(STORAGE_PREFIX + mode);
    return raw ? (JSON.parse(raw) as string[]) : [];
  } catch {
    return [];
  }
}

function saveHistory(mode: string, items: string[]) {
  if (!browser) return;
  localStorage.setItem(STORAGE_PREFIX + mode, JSON.stringify(items));
}

class SearchHistory {
  private histories = $state<Record<string, string[]>>({});

  getHistory(mode: string): string[] {
    if (!this.histories[mode]) {
      this.histories[mode] = loadHistory(mode);
    }
    return this.histories[mode];
  }

  addSearch(mode: string, term: string) {
    const t = term.trim();
    if (!t) return;
    const list = this.getHistory(mode).filter((item) => item !== t);
    list.unshift(t);
    if (list.length > MAX_HISTORY) list.length = MAX_HISTORY;
    this.histories[mode] = list;
    saveHistory(mode, list);
  }

  removeEntry(mode: string, term: string) {
    const list = this.getHistory(mode).filter((item) => item !== term);
    this.histories[mode] = list;
    saveHistory(mode, list);
  }

  clearHistory(mode: string) {
    this.histories[mode] = [];
    saveHistory(mode, []);
  }
}

export const searchHistory = new SearchHistory();
