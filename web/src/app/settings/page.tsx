"use client";

import { useEffect, useState } from "react";
import { defaultSettings, loadSettings, saveSettings } from "@/lib/settings";
import type { ApiSettings } from "@/lib/types";
import { api } from "@/lib/api";
import { Alert, Button, Card, Input, Label } from "@/components/ui";

export default function SettingsPage() {
  const [settings, setSettings] = useState<ApiSettings>(defaultSettings);
  const [saved, setSaved] = useState(false);
  const [testResult, setTestResult] = useState<string | null>(null);

  useEffect(() => {
    setSettings(loadSettings());
  }, []);

  function onSave(e: React.FormEvent) {
    e.preventDefault();
    saveSettings(settings);
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  }

  async function testConnection() {
    saveSettings(settings);
    try {
      const h = await api.health();
      setTestResult(`Connected — ${h.service} is ${h.status}`);
    } catch {
      setTestResult("Could not reach API. Is the backend running on this URL?");
    }
  }

  return (
    <div className="mx-auto max-w-lg space-y-6 animate-in">
      <header>
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="mt-1 text-[var(--text-muted)]">
          Configure how this console talks to your Go banking API.
        </p>
      </header>

      <Card>
        <form onSubmit={onSave} className="space-y-4">
          <div>
            <Label htmlFor="url">API base URL</Label>
            <Input
              id="url"
              placeholder="/api"
              value={settings.baseUrl}
              onChange={(e) => setSettings({ ...settings, baseUrl: e.target.value })}
            />
          </div>
          <div>
            <Label htmlFor="key">API key (X-API-Key)</Label>
            <Input
              id="key"
              type="password"
              value={settings.apiKey}
              onChange={(e) => setSettings({ ...settings, apiKey: e.target.value })}
              placeholder="Only required if AUTH_ENABLED=true"
            />
          </div>
          <div>
            <Label htmlFor="actor">Actor (X-Actor)</Label>
            <Input
              id="actor"
              value={settings.actor}
              onChange={(e) => setSettings({ ...settings, actor: e.target.value })}
            />
            <p className="mt-1 text-xs text-[var(--text-muted)]">Written to the audit log on transfers.</p>
          </div>

          {saved && <Alert tone="success">Settings saved to this browser.</Alert>}
          {testResult && (
            <Alert tone={testResult.startsWith("Connected") ? "success" : "danger"}>
              {testResult}
            </Alert>
          )}

          <div className="flex gap-2">
            <Button type="submit" className="flex-1">
              Save
            </Button>
            <Button type="button" variant="secondary" onClick={testConnection}>
              Test connection
            </Button>
          </div>
        </form>
      </Card>
    </div>
  );
}
