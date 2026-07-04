"use client";

import { useEffect, useState } from "react";
import {
  CheckCircle2,
  Globe,
  Key,
  Loader2,
  Server,
  Settings,
  Shield,
  User,
  Wifi,
  WifiOff,
} from "lucide-react";
import { defaultSettings, loadSettings, saveSettings } from "@/lib/settings";
import type { ApiSettings } from "@/lib/types";
import { api } from "@/lib/api";
import { Alert, Button, Card, Divider, Input, Label } from "@/components/ui";
import { cn } from "@/lib/utils";

export default function SettingsPage() {
  const [settings, setSettings] = useState<ApiSettings>(defaultSettings);
  const [saved, setSaved] = useState(false);
  const [testResult, setTestResult] = useState<"idle" | "testing" | "success" | "error">("idle");
  const [testMessage, setTestMessage] = useState<string>("");

  useEffect(() => {
    setSettings(loadSettings());
  }, []);

  function onSave(e: React.FormEvent) {
    e.preventDefault();
    saveSettings(settings);
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  }

  async function testConnection() {
    saveSettings(settings);
    setTestResult("testing");
    try {
      const h = await api.health();
      setTestResult("success");
      setTestMessage(`${h.service} — ${h.status}`);
    } catch {
      setTestResult("error");
      setTestMessage("Could not reach API. Is the backend running?");
    }
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6 animate-in">
      <header>
        <div className="flex items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-[var(--bg-surface)] text-[var(--text-muted)]">
            <Settings className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-xl font-bold tracking-tight md:text-2xl">Settings</h1>
            <p className="text-sm text-[var(--text-muted)]">
              Configure the API connection and authentication
            </p>
          </div>
        </div>
      </header>

      {/* Connection Status */}
      <div
        className={cn(
          "flex items-center gap-4 rounded-2xl border p-5 transition-colors",
          testResult === "success"
            ? "border-emerald-500/20 bg-emerald-950/20"
            : testResult === "error"
              ? "border-red-500/20 bg-red-950/20"
              : "border-[var(--border)] bg-[var(--bg-card)]",
        )}
      >
        <div
          className={cn(
            "flex h-10 w-10 items-center justify-center rounded-xl",
            testResult === "success"
              ? "bg-emerald-500/10 text-emerald-400"
              : testResult === "error"
                ? "bg-red-500/10 text-red-400"
                : "bg-[var(--bg-surface)] text-[var(--text-muted)]",
          )}
        >
          {testResult === "testing" ? (
            <Loader2 className="h-5 w-5 animate-spin" />
          ) : testResult === "success" ? (
            <Wifi className="h-5 w-5" />
          ) : testResult === "error" ? (
            <WifiOff className="h-5 w-5" />
          ) : (
            <Server className="h-5 w-5" />
          )}
        </div>
        <div className="flex-1">
          <p className="text-sm font-semibold">
            {testResult === "testing"
              ? "Testing connection..."
              : testResult === "success"
                ? "Connected"
                : testResult === "error"
                  ? "Connection Failed"
                  : "API Connection"}
          </p>
          <p className="text-xs text-[var(--text-muted)]">
            {testResult === "idle"
              ? `Target: ${settings.baseUrl || "/api"}`
              : testMessage}
          </p>
        </div>
        <Button
          variant="secondary"
          size="sm"
          onClick={testConnection}
          disabled={testResult === "testing"}
        >
          {testResult === "testing" ? "Testing…" : "Test"}
        </Button>
      </div>

      <Card title="API Configuration" subtitle="How this console connects to the Go banking backend">
        <form onSubmit={onSave} className="space-y-5">
          <div>
            <Label htmlFor="url">
              <Globe className="h-3.5 w-3.5" />
              API Base URL
            </Label>
            <Input
              id="url"
              placeholder="/api"
              value={settings.baseUrl}
              onChange={(e) => setSettings({ ...settings, baseUrl: e.target.value })}
            />
            <p className="mt-1.5 text-[11px] text-[var(--text-muted)]">
              Use <code className="rounded bg-[var(--bg-surface)] px-1 py-0.5 font-mono text-[10px]">/api</code> for
              Next.js proxy (recommended) or <code className="rounded bg-[var(--bg-surface)] px-1 py-0.5 font-mono text-[10px]">http://localhost:8080</code> for
              direct connection
            </p>
          </div>

          <Divider />

          <div>
            <Label htmlFor="key">
              <Key className="h-3.5 w-3.5" />
              API Key
            </Label>
            <Input
              id="key"
              type="password"
              value={settings.apiKey}
              onChange={(e) => setSettings({ ...settings, apiKey: e.target.value })}
              placeholder="Only required if AUTH_ENABLED=true"
            />
            <p className="mt-1.5 text-[11px] text-[var(--text-muted)]">
              Sent as <code className="rounded bg-[var(--bg-surface)] px-1 py-0.5 font-mono text-[10px]">X-API-Key</code> header
              on all <code className="rounded bg-[var(--bg-surface)] px-1 py-0.5 font-mono text-[10px]">/v1/*</code> requests
            </p>
          </div>

          <div>
            <Label htmlFor="actor">
              <User className="h-3.5 w-3.5" />
              Actor Identity
            </Label>
            <Input
              id="actor"
              value={settings.actor}
              onChange={(e) => setSettings({ ...settings, actor: e.target.value })}
              placeholder="web-console"
            />
            <p className="mt-1.5 text-[11px] text-[var(--text-muted)]">
              Written to the audit log on every transfer as the originating user
            </p>
          </div>

          {saved && (
            <Alert tone="success">
              <div className="flex items-center gap-2">
                <CheckCircle2 className="h-4 w-4" />
                Settings saved to browser localStorage
              </div>
            </Alert>
          )}

          <div className="flex gap-3 pt-2">
            <Button type="submit" className="flex-1" size="lg">
              <CheckCircle2 className="h-4 w-4" />
              Save Settings
            </Button>
            <Button
              type="button"
              variant="secondary"
              onClick={() => {
                setSettings(defaultSettings);
                saveSettings(defaultSettings);
                setSaved(true);
                setTimeout(() => setSaved(false), 3000);
              }}
            >
              Reset to Defaults
            </Button>
          </div>
        </form>
      </Card>

      {/* Info Card */}
      <Card className="border-blue-500/10 bg-blue-950/10">
        <div className="flex gap-3">
          <Shield className="mt-0.5 h-5 w-5 shrink-0 text-blue-400" />
          <div>
            <p className="text-sm font-medium text-blue-300">About Security</p>
            <p className="mt-1 text-xs text-blue-300/70 leading-relaxed">
              Settings are stored in your browser&apos;s localStorage. The API key is never sent to any
              server other than your configured backend URL. When using the{" "}
              <code className="rounded bg-blue-500/10 px-1 py-0.5 font-mono text-[10px]">/api</code>{" "}
              proxy, requests are proxied through the Next.js server to avoid CORS issues.
            </p>
          </div>
        </div>
      </Card>
    </div>
  );
}
