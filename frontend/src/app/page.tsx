import Link from "next/link";
import { Shield, Lock, Activity, CheckCircle, ArrowRight, Server, Terminal, Radio } from "lucide-react";

export default function DashboardPage() {
  return (
    <div className="space-y-8">
      {/* Top Hero / Status Bar */}
      <div className="p-6 rounded-xl bg-vulnex-card border border-vulnex-border flex flex-col md:flex-row justify-between items-start md:items-center gap-4 shadow-lg">
        <div>
          <div className="flex items-center space-x-2">
            <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse mr-1.5"></span>
              Core System Healthy
            </span>
            <span className="text-xs text-slate-400 font-mono">v0.1.0-dev</span>
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-white mt-2">
            Defensive Security Overview
          </h1>
          <p className="text-sm text-slate-400 mt-1 max-w-2xl">
            VULNEX platform manages authorized vulnerability assessments, isolated scanner execution (Nmap, OWASP ZAP), and normalized risk scoring.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Link
            href="/scans"
            className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-medium rounded-lg transition flex items-center space-x-2 shadow-sm"
          >
            <Activity className="w-4 h-4" />
            <span>Launch Assessment</span>
          </Link>
          <Link
            href="/workspaces"
            className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-200 text-sm font-medium rounded-lg border border-slate-700 transition"
          >
            Manage Scope
          </Link>
        </div>
      </div>

      {/* KPI Stats Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="p-5 rounded-xl bg-vulnex-card border border-vulnex-border">
          <div className="flex items-center justify-between text-slate-400 text-xs font-medium uppercase tracking-wider">
            <span>Critical Findings</span>
            <span className="p-1.5 rounded bg-red-500/10 text-red-400"><Lock className="w-4 h-4" /></span>
          </div>
          <div className="text-2xl font-bold text-white mt-2">0</div>
          <div className="text-xs text-emerald-400 mt-1">No unmitigated critical risks</div>
        </div>

        <div className="p-5 rounded-xl bg-vulnex-card border border-vulnex-border">
          <div className="flex items-center justify-between text-slate-400 text-xs font-medium uppercase tracking-wider">
            <span>High / Medium Risks</span>
            <span className="p-1.5 rounded bg-amber-500/10 text-amber-400"><Activity className="w-4 h-4" /></span>
          </div>
          <div className="text-2xl font-bold text-white mt-2">0</div>
          <div className="text-xs text-slate-400 mt-1">Ready for scan ingestion</div>
        </div>

        <div className="p-5 rounded-xl bg-vulnex-card border border-vulnex-border">
          <div className="flex items-center justify-between text-slate-400 text-xs font-medium uppercase tracking-wider">
            <span>Authorized Scopes</span>
            <span className="p-1.5 rounded bg-blue-500/10 text-blue-400"><CheckCircle className="w-4 h-4" /></span>
          </div>
          <div className="text-2xl font-bold text-white mt-2">1 Active</div>
          <div className="text-xs text-blue-400 mt-1">Strict RFC allowlists active</div>
        </div>

        <div className="p-5 rounded-xl bg-vulnex-card border border-vulnex-border">
          <div className="flex items-center justify-between text-slate-400 text-xs font-medium uppercase tracking-wider">
            <span>Scanner Runtimes</span>
            <span className="p-1.5 rounded bg-purple-500/10 text-purple-400"><Server className="w-4 h-4" /></span>
          </div>
          <div className="text-2xl font-bold text-white mt-2">2 Ready</div>
          <div className="text-xs text-purple-400 mt-1">Nmap 7.94 & OWASP ZAP 2.14</div>
        </div>
      </div>

      {/* Security Policies and Architecture Highlights */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="p-6 rounded-xl bg-vulnex-card border border-vulnex-border space-y-4">
          <h2 className="text-lg font-semibold text-white flex items-center space-x-2">
            <Shield className="w-5 h-5 text-blue-400" />
            <span>Defensive Architecture Guardrails</span>
          </h2>
          <ul className="space-y-3 text-sm text-slate-300">
            <li className="flex items-start space-x-3">
              <CheckCircle className="w-4 h-4 text-emerald-400 mt-0.5 flex-shrink-0" />
              <div>
                <strong className="text-white">Authorization-First Guarantee:</strong> No network packet or probe is transmitted without active, non-expired, written client authorization (<code className="text-xs text-blue-400">ADR-006</code>).
              </div>
            </li>
            <li className="flex items-start space-x-3">
              <CheckCircle className="w-4 h-4 text-emerald-400 mt-0.5 flex-shrink-0" />
              <div>
                <strong className="text-white">SSRF & Loopback Hard-Blocks:</strong> Cloud metadata (<code className="text-xs text-slate-300 font-mono">169.254.169.254</code>), loopbacks, link-local, and internal subnets are unconditionally rejected.
              </div>
            </li>
            <li className="flex items-start space-x-3">
              <CheckCircle className="w-4 h-4 text-emerald-400 mt-0.5 flex-shrink-0" />
              <div>
                <strong className="text-white">Workspace Private Scanning Policy:</strong> Testing internal RFC 1918 subnets requires explicit workspace administrator opt-in and audit logs (<code className="text-xs text-blue-400">ADR-015</code>).
              </div>
            </li>
            <li className="flex items-start space-x-3">
              <CheckCircle className="w-4 h-4 text-emerald-400 mt-0.5 flex-shrink-0" />
              <div>
                <strong className="text-white">Bounded Stream Parsing:</strong> Scanner outputs are streamed through hard-capped 50MB readers, preventing memory exhaustion attacks.
              </div>
            </li>
          </ul>
        </div>

        <div className="p-6 rounded-xl bg-vulnex-card border border-vulnex-border space-y-4">
          <h2 className="text-lg font-semibold text-white flex items-center space-x-2">
            <Terminal className="w-5 h-5 text-emerald-400" />
            <span>Quick Start & Control Plane</span>
          </h2>
          <p className="text-sm text-slate-400">
            Initial setup requires the cryptographically secure setup token generated by the VULNEX installer CLI:
          </p>

          <div className="p-3 rounded-lg bg-black/60 border border-slate-800 font-mono text-xs text-slate-300 space-y-1">
            <div className="text-slate-500"># Verify system and generate admin bootstrap token</div>
            <div className="text-emerald-400">$ ./bin/vulnex install --mode server</div>
            <div className="text-slate-500 mt-2"># View real-time status and logs</div>
            <div className="text-blue-400">$ ./bin/vulnex status</div>
          </div>

          <div className="pt-2">
            <Link
              href="/setup"
              className="inline-flex items-center space-x-2 text-sm text-blue-400 hover:text-blue-300 font-medium transition"
            >
              <span>Proceed to Administrator Onboarding</span>
              <ArrowRight className="w-4 h-4" />
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
