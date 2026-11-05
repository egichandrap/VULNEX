"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { api } from "@/lib/api";
import { Globe, Play, CheckCircle, Clock, AlertTriangle, ShieldAlert, LogIn } from "lucide-react";

export default function ScansPage() {
  const [target, setTarget] = useState("");
  const [scannerProfile, setScannerProfile] = useState("nmap-top-ports");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<{ scan_id: string; status: string } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState<boolean | null>(null);

  useEffect(() => {
    if (typeof window !== "undefined") {
      setIsAuthenticated(!!localStorage.getItem("vulnex_token"));
    }
  }, []);

  const handleLaunchScan = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setResult(null);

    if (!target.trim()) return;

    if (typeof window !== "undefined" && !localStorage.getItem("vulnex_token")) {
      setError("Authentication required: Silakan Sign In terlebih dahulu untuk menjalankan scanning keamanan.");
      return;
    }

    setLoading(true);

    try {
      const res = await api.requestScan("00000000-0000-0000-0000-000000000002", {
        workspace_id: "00000000-0000-0000-0000-000000000001",
        authorization_id: "00000000-0000-0000-0000-000000000003",
        profile_version_id: "00000000-0000-0000-0000-000000000004",
        idempotency_key: `scan-${Date.now()}`,
        targets: [target.trim()],
      });

      setResult(res);
    } catch (err: any) {
      setError(err.message || "Failed to launch scan.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-white flex items-center space-x-2">
          <Globe className="w-6 h-6 text-blue-400" />
          <span>Security Scanner Execution</span>
        </h1>
        <p className="text-sm text-slate-400 mt-1">
          Execute authorized security discovery and baseline vulnerability assessment runs.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Launch Form */}
        <div className="lg:col-span-1 p-6 rounded-xl bg-vulnex-card border border-vulnex-border space-y-4">
          <h2 className="text-lg font-bold text-white">Trigger Assessment</h2>

          {isAuthenticated === false && (
            <div className="p-3 rounded-lg bg-amber-500/10 border border-amber-500/30 text-amber-300 text-xs flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <ShieldAlert className="w-4 h-4 flex-shrink-0" />
                <span>Login diperlukan sebelum scanning</span>
              </div>
              <Link
                href="/login"
                className="px-2.5 py-1 bg-amber-500 hover:bg-amber-400 text-slate-900 font-semibold rounded text-[11px] transition"
              >
                Sign In
              </Link>
            </div>
          )}

          {error && (
            <div className="p-3 rounded-lg bg-red-500/10 border border-red-500/30 text-red-400 text-xs flex items-start space-x-2">
              <AlertTriangle className="w-4 h-4 flex-shrink-0 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          {result && (
            <div className="p-4 rounded-lg bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-xs space-y-2">
              <div className="flex items-start space-x-2">
                <CheckCircle className="w-5 h-5 flex-shrink-0 mt-0.5 text-emerald-400" />
                <div className="space-y-1">
                  <div className="font-bold text-sm text-white">Scan Berhasil Diantrekan!</div>
                  <div className="font-mono text-xs text-slate-300">
                    ID: <span className="text-emerald-300 font-semibold">{result.scan_id || (result as any).ScanID}</span>
                  </div>
                  <div className="text-[11px] text-slate-400">
                    Status: <span className="font-semibold text-blue-400">{result.status || (result as any).Status || "QUEUED"}</span>
                  </div>
                </div>
              </div>
              <div className="pt-2 border-t border-emerald-500/20 flex items-center justify-between">
                <span className="text-slate-300 text-[11px]">Lihat hasil pemindaian:</span>
                <Link
                  href="/findings"
                  className="px-3 py-1 bg-emerald-600 hover:bg-emerald-500 text-white font-semibold rounded text-xs transition inline-flex items-center space-x-1"
                >
                  <span>Buka Findings &rarr;</span>
                </Link>
              </div>
            </div>
          )}

          <form onSubmit={handleLaunchScan} className="space-y-4">
            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1">
                Target (Domain, IP, or Web URL)
              </label>
              <input
                type="text"
                required
                value={target}
                onChange={(e) => setTarget(e.target.value)}
                placeholder="e.g. scanme.nmap.org or https://target.app"
                className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-lg text-sm text-white focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>

            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1">
                Approved Scanner Profile
              </label>
              <select
                value={scannerProfile}
                onChange={(e) => setScannerProfile(e.target.value)}
                className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-lg text-sm text-white focus:outline-none focus:border-blue-500"
              >
                <option value="nmap-top-ports">Nmap — Top Ports Discovery (TCP)</option>
                <option value="nmap-discovery">Nmap — Fast Ping Sweep</option>
                <option value="zap-baseline">OWASP ZAP — Passive Web Baseline</option>
              </select>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full py-2.5 px-4 bg-blue-600 hover:bg-blue-500 disabled:bg-slate-700 text-white font-medium rounded-lg text-sm transition shadow flex items-center justify-center space-x-2"
            >
              <Play className="w-4 h-4 fill-current" />
              <span>{loading ? "Queueing Task..." : "Queue Assessment Task"}</span>
            </button>
          </form>
        </div>

        {/* Real-time Execution Feed & Invariants */}
        <div className="lg:col-span-2 space-y-4">
          <div className="p-6 rounded-xl bg-vulnex-card border border-vulnex-border space-y-4">
            <h2 className="text-lg font-bold text-white flex items-center justify-between">
              <span>Scanner Execution Queue (SKIP LOCKED)</span>
              <span className="text-xs font-normal text-slate-400">Worker Daemon: Active</span>
            </h2>

            <div className="space-y-3">
              <div className="p-4 rounded-lg bg-slate-900/60 border border-slate-800 flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div className="p-2 rounded bg-emerald-500/10 text-emerald-400">
                    <CheckCircle className="w-5 h-5" />
                  </div>
                  <div>
                    <div className="text-sm font-semibold text-white">Baseline Assessment Conformance Run</div>
                    <div className="text-xs text-slate-400 font-mono">Profile: zap-baseline • Targets: 1 URL</div>
                  </div>
                </div>
                <span className="px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  COMPLETED
                </span>
              </div>

              <div className="p-4 rounded-lg bg-slate-900/60 border border-slate-800 flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div className="p-2 rounded bg-blue-500/10 text-blue-400">
                    <Clock className="w-5 h-5" />
                  </div>
                  <div>
                    <div className="text-sm font-semibold text-white">Network Perimeter Discovery Sweep</div>
                    <div className="text-xs text-slate-400 font-mono">Profile: nmap-top-ports • Targets: 1 Host</div>
                  </div>
                </div>
                <span className="px-2.5 py-1 rounded-full text-xs font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20">
                  QUEUED
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
