"use client";

import { useEffect, useState } from "react";
import { api, WorkspaceDTO } from "@/lib/api";
import { Layers, Plus, ShieldCheck, ShieldAlert, AlertCircle, CheckCircle2 } from "lucide-react";

export default function WorkspacesPage() {
  const [workspaces, setWorkspaces] = useState<WorkspaceDTO[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Modal / form states
  const [newWsName, setNewWsName] = useState("");
  const [creating, setCreating] = useState(false);

  // Policy toggle state
  const [selectedWs, setSelectedWs] = useState<WorkspaceDTO | null>(null);
  const [policyJustification, setPolicyJustification] = useState("");
  const [policyLoading, setPolicyLoading] = useState(false);
  const [policyMsg, setPolicyMsg] = useState<string | null>(null);

  const fetchWorkspaces = async () => {
    try {
      setLoading(true);
      const res = await api.listWorkspaces();
      setWorkspaces(res.workspaces || []);
    } catch (err: any) {
      setError(err.message || "Failed to load workspaces. Please sign in.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchWorkspaces();
  }, []);

  const handleCreateWorkspace = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newWsName.trim()) return;

    try {
      setCreating(true);
      await api.createWorkspace(newWsName.trim());
      setNewWsName("");
      await fetchWorkspaces();
    } catch (err: any) {
      alert(err.message || "Failed to create workspace");
    } finally {
      setCreating(false);
    }
  };

  const handleTogglePrivateNetwork = async (ws: WorkspaceDTO, newSetting: boolean) => {
    if (newSetting && !policyJustification.trim()) {
      alert("A documented justification is mandatory to enable private network scanning (ADR-015).");
      return;
    }

    try {
      setPolicyLoading(true);
      await api.updateWorkspacePolicy(
        ws.id,
        newSetting,
        newSetting ? policyJustification.trim() : "Disabled by administrator"
      );

      setPolicyMsg(`Workspace private scanning policy updated to ${newSetting ? 'ENABLED' : 'DISABLED'}`);
      setSelectedWs(null);
      setPolicyJustification("");
      await fetchWorkspaces();
    } catch (err: any) {
      alert(err.message || "Failed to update network policy");
    } finally {
      setPolicyLoading(false);
    }
  };

  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center space-x-2">
            <Layers className="w-6 h-6 text-blue-400" />
            <span>Workspaces & Tenancy Scope</span>
          </h1>
          <p className="text-sm text-slate-400 mt-1">
            Manage multi-tenant assessment environments and network scanning safety policies.
          </p>
        </div>

        <form onSubmit={handleCreateWorkspace} className="flex items-center space-x-2">
          <input
            type="text"
            required
            placeholder="New Workspace Name"
            value={newWsName}
            onChange={(e) => setNewWsName(e.target.value)}
            className="px-3 py-1.5 bg-slate-900 border border-slate-700 rounded-lg text-sm text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
          />
          <button
            type="submit"
            disabled={creating}
            className="px-3 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-sm font-medium transition flex items-center space-x-1"
          >
            <Plus className="w-4 h-4" />
            <span>Create</span>
          </button>
        </form>
      </div>

      {policyMsg && (
        <div className="p-4 rounded-lg bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-sm flex items-center space-x-2">
          <CheckCircle2 className="w-5 h-5 flex-shrink-0" />
          <span>{policyMsg}</span>
        </div>
      )}

      {error && (
        <div className="p-4 rounded-lg bg-amber-500/10 border border-amber-500/30 text-amber-400 text-sm flex items-center space-x-2">
          <AlertCircle className="w-5 h-5 flex-shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* Workspaces List */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {workspaces.map((ws) => (
          <div
            key={ws.id}
            className="p-6 rounded-xl bg-vulnex-card border border-vulnex-border flex flex-col justify-between space-y-4 hover:border-slate-700 transition"
          >
            <div>
              <div className="flex items-center justify-between">
                <span className="text-xs font-mono text-slate-500">ID: {ws.id.slice(0, 8)}...</span>
                <span className="px-2 py-0.5 rounded text-[10px] font-semibold tracking-wider bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  {ws.status}
                </span>
              </div>
              <h2 className="text-lg font-bold text-white mt-2">{ws.name}</h2>
            </div>

            {/* Policy Section (ADR-015) */}
            <div className="pt-4 border-t border-slate-800/80 space-y-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-1.5 text-xs text-slate-300">
                  {ws.allow_private_network_scan ? (
                    <ShieldAlert className="w-4 h-4 text-amber-400" />
                  ) : (
                    <ShieldCheck className="w-4 h-4 text-emerald-400" />
                  )}
                  <span>Private Network Scanning:</span>
                </div>
                <span className={`text-xs font-bold ${ws.allow_private_network_scan ? 'text-amber-400' : 'text-slate-400'}`}>
                  {ws.allow_private_network_scan ? 'ENABLED' : 'DENIED'}
                </span>
              </div>

              {selectedWs?.id === ws.id ? (
                <div className="p-3 rounded-lg bg-black/40 border border-slate-800 space-y-2 text-xs">
                  <label className="block text-slate-300 font-medium">Documented Justification (ADR-015):</label>
                  <input
                    type="text"
                    required
                    placeholder="e.g. Authorized internal audit staging environment"
                    value={policyJustification}
                    onChange={(e) => setPolicyJustification(e.target.value)}
                    className="w-full px-2 py-1 bg-slate-900 border border-slate-700 rounded text-white focus:outline-none focus:border-blue-500"
                  />
                  <div className="flex items-center space-x-2 pt-1">
                    <button
                      onClick={() => handleTogglePrivateNetwork(ws, true)}
                      disabled={policyLoading}
                      className="px-2 py-1 bg-amber-600 hover:bg-amber-500 text-white rounded font-medium"
                    >
                      Confirm Enable
                    </button>
                    <button
                      onClick={() => setSelectedWs(null)}
                      className="px-2 py-1 bg-slate-800 text-slate-400 hover:text-white rounded"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              ) : (
                <button
                  onClick={() => {
                    if (ws.allow_private_network_scan) {
                      handleTogglePrivateNetwork(ws, false);
                    } else {
                      setSelectedWs(ws);
                    }
                  }}
                  className={`w-full py-1.5 px-3 rounded text-xs font-medium border transition ${
                    ws.allow_private_network_scan
                      ? 'border-slate-700 hover:bg-slate-800 text-slate-300'
                      : 'border-amber-500/30 bg-amber-500/10 hover:bg-amber-500/20 text-amber-400'
                  }`}
                >
                  {ws.allow_private_network_scan ? 'Disable Private Scanning' : 'Enable Private Scanning (ADR-015)'}
                </button>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
