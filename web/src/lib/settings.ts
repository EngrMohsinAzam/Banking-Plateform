import type { ApiSettings } from "./types";

const STORAGE_KEY = "banking_api_settings";

export const defaultSettings: ApiSettings = {
  baseUrl: process.env.NEXT_PUBLIC_API_URL ?? "/api",
  apiKey: process.env.NEXT_PUBLIC_API_KEY ?? "",
  actor: "web-console",
};

export function loadSettings(): ApiSettings {
  if (typeof window === "undefined") return defaultSettings;
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return defaultSettings;
    const saved = JSON.parse(raw) as Partial<ApiSettings>;
    const merged = { ...defaultSettings, ...saved };
    // Migrate old direct-backend URL to Next.js proxy (avoids CORS).
    if (
      merged.baseUrl === "http://localhost:8080" ||
      merged.baseUrl === "http://127.0.0.1:8080"
    ) {
      merged.baseUrl = defaultSettings.baseUrl;
    }
    return merged;
  } catch {
    return defaultSettings;
  }
}

export function saveSettings(settings: ApiSettings): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
}
