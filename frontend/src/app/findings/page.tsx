"use client";

import { useState } from "react";
import { AlertTriangle, ShieldCheck, CheckCircle2, Filter, AlertOctagon } from "lucide-react";

export default function FindingsPage() {
  const [filter, setFilter] = useState("ALL");

  const mockFindings = [
    {
      id: "fnd-001",
      title: "Content Security Policy (CSP) Header Not Set",
      description: "Content Security Policy is an effective defense against Cross-Site Scripting (XSS) and data injection attacks.",
      severity: "MEDIUM",
      confidence: "HIGH",
      status: "OPEN",
      source_scanner: "zap",
      risk_score: 5.4,
      target: "https://target.vulnex.local/login",
    },
    {
      id: "fnd-002",
      title: "Open Port 22 (SSH)",
      description: "Exposed SSH administrative port on external interface ECS Bastion.",
      severity: "LOW",
      confidence: "CONFIRMED",
      status: "OPEN",
      source_scanner: "nmap",
      risk_score: 2.5,
      target: "tcp/22 (93.184.216.34)",
    },
    {
      id: "fnd-003",
      title: "Anti-CSRF Tokens Check Missing",
      description: "State-changing form without anti-CSRF token verification.",
      severity: "HIGH",
      confidence: "HIGH",
      status: "RESOLVED",
      source_scanner: "zap",
      risk_score: 7.2,
      target: "https://target.vulnex.local/transfer",
    },
  ];

  const getSeverityBadge = (severity: string) => {
    switch (severity) {
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

        <div className="flex items-center space-x-2 bg-vulnex-card border border-vulnex-border rounded-lg p-1 text-xs">
          <button
            onClick={() => setFilter("ALL")}
            className={`px-3 py-1.5 rounded-md font-medium transition ${
              filter === "ALL" ? "bg-blue-600 text-white" : "text-slate-400 hover:text-white"
            }`}
          >
            All Findings
          </button>
          <button
            onClick={() => setFilter("OPEN")}
            className={`px-3 py-1.5 rounded-md font-medium transition ${
              filter === "OPEN" ? "bg-blue-600 text-white" : "text-slate-400 hover:text-white"
            }`}
          >
            Open Only
          </button>
          <button
            onClick={() => setFilter("RESOLVED")}
            className={`px-3 py-1.5 rounded-md font-medium transition ${
              filter === "RESOLVED" ? "bg-blue-600 text-white" : "text-slate-400 hover:text-white"
            }`}
          >
            Resolved
          </button>
        </div>
      </div>

      {/* Findings Table */}
      <div className="rounded-xl bg-vulnex-card border border-vulnex-border overflow-hidden">
        <table className="w-full text-left text-sm text-slate-300">
          <thead className="bg-slate-900/80 text-xs font-semibold uppercase text-slate-400 border-b border-vulnex-border">
            <tr>
              <th className="py-3 px-4">Severity & Score</th>
              <th className="py-3 px-4">Title & Target</th>
              <th className="py-3 px-4">Scanner</th>
              <th className="py-3 px-4">Status</th>
              <th className="py-3 px-4 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800/60">
            {mockFindings.map((f) => (
              <tr key={f.id} className="hover:bg-slate-800/40 transition">
                <td className="py-4 px-4 whitespace-nowrap">
                  <div className="flex items-center space-x-2">
                    <span className={`px-2 py-0.5 rounded text-[11px] font-bold border ${getSeverityBadge(f.severity)}`}>
                      {f.severity}
                    </span>
                    <span className="font-mono text-xs font-semibold text-white">
                      {f.risk_score.toFixed(1)}/10
                    </span>
                  </div>
                </td>
                <td className="py-4 px-4">
                  <div className="font-semibold text-white">{f.title}</div>
                  <div className="text-xs text-slate-400 font-mono mt-0.5">{f.target}</div>
                </td>
                <td className="py-4 px-4 whitespace-nowrap">
                  <span className="text-xs font-mono uppercase bg-slate-800 px-2 py-1 rounded text-slate-300 border border-slate-700">
                    {f.source_scanner}
                  </span>
                </td>
                <td className="py-4 px-4 whitespace-nowrap">
                  <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium border ${
                    f.status === 'RESOLVED'
                      ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                      : 'bg-amber-500/10 text-amber-400 border-amber-500/20'
                  }`}>
                    {f.status}
                  </span>
                </td>
                <td className="py-4 px-4 text-right whitespace-nowrap">
                  <button
                    onClick={() => alert(`Triage action for ${f.id}`)}
                    className="text-xs text-blue-400 hover:text-blue-300 font-medium px-2.5 py-1 rounded border border-blue-500/20 hover:border-blue-500/40 transition"
                  >
                    Triage
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
