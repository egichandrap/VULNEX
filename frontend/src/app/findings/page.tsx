"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { api, FindingDTO } from "@/lib/api";
import { AlertTriangle, ShieldCheck, CheckCircle2, Filter, AlertOctagon, RefreshCw, ShieldAlert, ExternalLink } from "lucide-react";

export default function FindingsPage() {
  const [filter, setFilter] = useState("ALL");
  const [findings, setFindings] = useState<FindingDTO[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchFindings = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.listFindings("00000000-0000-0000-0000-000000000002");
      setFindings(res.findings || []);
    } catch (err: any) {
      setError(err.message || "Failed to load findings.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchFindings();
  }, []);

  const handleTriage = async (findingId: string, currentStatus: string) => {
    const newStatus = currentStatus === "OPEN" ? "RESOLVED" : "OPEN";
    try {
      await api.triageFinding(findingId, newStatus, "Operator triage action via web console");
      fetchFindings();
    } catch (err: any) {
      alert(`Triage failed: ${err.message}`);
    }
  };

  const getSeverityBadge = (severity: string) => {
    switch (severity.toUpperCase()) {
      case "CRITICAL":
        return "bg-red-500/10 text-red-400 border-red-500/30";
      case "HIGH":
        return "bg-orange-500/10 text-orange-400 border-orange-500/30";
      case "MEDIUM":
        return "bg-amber-500/10 text-amber-400 border-amber-500/30";
      case "LOW":
        return "bg-blue-500/10 text-blue-400 border-blue-500/30";
      default:
        return "bg-slate-500/10 text-slate-400 border-slate-500/30";
    }
  };

  const filteredFindings = findings.filter((f) => {
    if (filter === "ALL") return true;
    return f.status === filter;
  });

  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center space-x-2">
            <AlertTriangle className="w-6 h-6 text-amber-400" />
            <span>Vulnerability Findings & Triage</span>
          </h1>
          <p className="text-sm text-slate-400 mt-1">
            Deduplicated security observations with explainable mathematical risk scoring.
          </p>
        </div>

        <div className="flex items-center space-x-3">
          <button
            onClick={fetchFindings}
            disabled={loading}
            className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 border border-slate-700 text-xs font-medium text-slate-300 transition"
            title="Refresh findings"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin" : ""}`} />
            <span>Refresh</span>
          </button>

          <div className="flex items-center space-x-1 bg-vulnex-card border border-vulnex-border rounded-lg p-1 text-xs">
            <button
              onClick={() => setFilter("ALL")}
              className={`px-3 py-1.5 rounded-md font-medium transition ${
                filter === "ALL" ? "bg-blue-600 text-white" : "text-slate-400 hover:text-white"
              }`}
            >
              All ({findings.length})
            </button>
            <button
              onClick={() => setFilter("OPEN")}
              className={`px-3 py-1.5 rounded-md font-medium transition ${
                filter === "OPEN" ? "bg-blue-600 text-white" : "text-slate-400 hover:text-white"
              }`}
            >
              Open ({findings.filter((f) => f.status === "OPEN").length})
            </button>
            <button
              onClick={() => setFilter("RESOLVED")}
              className={`px-3 py-1.5 rounded-md font-medium transition ${
                filter === "RESOLVED" ? "bg-blue-600 text-white" : "text-slate-400 hover:text-white"
              }`}
            >
              Resolved ({findings.filter((f) => f.status === "RESOLVED").length})
            </button>
          </div>
        </div>
      </div>

      {error && (
        <div className="p-4 rounded-lg bg-red-500/10 border border-red-500/30 text-red-400 text-xs flex items-center space-x-2">
          <ShieldAlert className="w-4 h-4 flex-shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {loading && findings.length === 0 ? (
        <div className="p-12 text-center bg-vulnex-card border border-vulnex-border rounded-xl space-y-2">
          <RefreshCw className="w-8 h-8 text-blue-400 mx-auto animate-spin" />
          <p className="text-sm text-slate-400">Loading findings repository...</p>
        </div>
      ) : filteredFindings.length === 0 ? (
        <div className="p-12 text-center bg-vulnex-card border border-vulnex-border rounded-xl space-y-3">
          <ShieldCheck className="w-12 h-12 text-emerald-400 mx-auto" />
          <h3 className="text-lg font-bold text-white">Belum Ada Temuan Kerentanan</h3>
          <p className="text-sm text-slate-400 max-w-md mx-auto">
            Tidak ada temuan yang cocok dengan filter ({filter}). Jalankan pemindaian di menu Scans untuk mendeteksi kerentanan target.
          </p>
          <div className="pt-2">
            <Link
              href="/scans"
              className="inline-flex items-center space-x-2 px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white font-medium rounded-lg text-sm transition"
            >
              <span>Buka Menu Scans</span>
              <ExternalLink className="w-4 h-4" />
            </Link>
          </div>
        </div>
      ) : (
        <div className="rounded-xl bg-vulnex-card border border-vulnex-border overflow-hidden">
          <table className="w-full text-left text-sm text-slate-300">
            <thead className="bg-slate-900/80 text-xs font-semibold uppercase text-slate-400 border-b border-vulnex-border">
              <tr>
                <th className="py-3 px-4">Severity & Score</th>
                <th className="py-3 px-4">Title & Description</th>
                <th className="py-3 px-4">Scanner</th>
                <th className="py-3 px-4">Status</th>
                <th className="py-3 px-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60">
              {filteredFindings.map((f) => (
                <tr key={f.id} className="hover:bg-slate-800/40 transition">
                  <td className="py-4 px-4 whitespace-nowrap align-top">
                    <div className="flex items-center space-x-2">
                      <span className={`px-2 py-0.5 rounded text-[11px] font-bold border ${getSeverityBadge(f.severity)}`}>
                        {f.severity}
                      </span>
                      <span className="font-mono text-xs font-semibold text-white">
                        {f.risk_score ? f.risk_score.toFixed(1) : "0.0"}/10
                      </span>
                    </div>
                    <div className="text-[10px] text-slate-400 mt-1">
                      Conf: <span className="font-semibold">{f.confidence}</span>
                    </div>
                  </td>
                  <td className="py-4 px-4 align-top">
                    <div className="font-semibold text-white">{f.title}</div>
                    <div className="text-xs text-slate-400 mt-1 line-clamp-2">{f.description}</div>
                    <div className="font-mono text-[10px] text-slate-500 mt-1">Scan ID: {f.scan_id}</div>
                  </td>
                  <td className="py-4 px-4 whitespace-nowrap align-top">
                    <span className="text-xs font-mono uppercase bg-slate-800 px-2 py-1 rounded text-slate-300 border border-slate-700">
                      {f.source_scanner}
                    </span>
                  </td>
                  <td className="py-4 px-4 whitespace-nowrap align-top">
                    <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium border ${
                      f.status === 'RESOLVED'
                        ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                        : 'bg-amber-500/10 text-amber-400 border-amber-500/20'
                    }`}>
                      {f.status}
                    </span>
                  </td>
                  <td className="py-4 px-4 text-right whitespace-nowrap align-top">
                    <button
                      onClick={() => handleTriage(f.id, f.status)}
                      className="text-xs text-blue-400 hover:text-blue-300 font-medium px-2.5 py-1 rounded border border-blue-500/20 hover:border-blue-500/40 transition"
                    >
                      {f.status === "OPEN" ? "Mark Resolved" : "Re-open"}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
