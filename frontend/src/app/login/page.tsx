"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";
import { Lock, LogIn, AlertCircle } from "lucide-react";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    const identifier = email.trim();
    const loginEmail = !identifier.includes("@") ? `${identifier}@vulnex.local` : identifier;

    try {
      await api.login({
        email: loginEmail,
        password,
      });

      router.push("/workspaces");
    } catch (err: any) {
      setError(err.message || "Invalid credentials.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-md mx-auto my-16 p-8 bg-vulnex-card border border-vulnex-border rounded-xl shadow-xl">
      <div className="flex items-center space-x-3 mb-6">
        <div className="p-2.5 rounded-lg bg-blue-600/20 text-blue-400 border border-blue-500/30">
          <Lock className="w-6 h-6" />
        </div>
        <div>
          <h1 className="text-xl font-bold text-white">Sign In to VULNEX</h1>
          <p className="text-xs text-slate-400">Authorized personnel only</p>
        </div>
      </div>

      {error && (
        <div className="mb-6 p-4 rounded-lg bg-red-500/10 border border-red-500/30 text-red-400 text-xs flex items-start space-x-2">
          <AlertCircle className="w-4 h-4 flex-shrink-0 mt-0.5" />
          <span>{error}</span>
        </div>
      )}

      <form onSubmit={handleLogin} className="space-y-4">
        <div>
          <label className="block text-xs font-medium text-slate-300 mb-1">
            Username or Email
          </label>
          <input
            type="text"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="admin or admin@vulnex.local"
            className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-lg text-sm text-white focus:outline-none focus:border-blue-500"
          />
        </div>

        <div>
          <label className="block text-xs font-medium text-slate-300 mb-1">
            Password
          </label>
          <input
            type="password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="admin"
            className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-lg text-sm text-white focus:outline-none focus:border-blue-500"
          />
        </div>

        <div className="p-2.5 rounded bg-blue-500/10 border border-blue-500/20 text-[11px] text-blue-300">
          Default credentials: <code className="font-mono bg-blue-900/40 px-1 py-0.5 rounded text-white">admin</code> / <code className="font-mono bg-blue-900/40 px-1 py-0.5 rounded text-white">admin</code> (dapat diubah setelah login).
        </div>

        <button
          type="submit"
          disabled={loading}
          className="w-full mt-4 py-2.5 px-4 bg-blue-600 hover:bg-blue-500 disabled:bg-slate-700 text-white font-medium rounded-lg text-sm transition shadow-md flex items-center justify-center space-x-2"
        >
          {loading ? (
            <span>Signing in...</span>
          ) : (
            <>
              <LogIn className="w-4 h-4" />
              <span>Authenticate</span>
            </>
          )}
        </button>
      </form>
    </div>
  );
}
