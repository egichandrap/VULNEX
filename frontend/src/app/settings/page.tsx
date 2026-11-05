"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";
import { KeyRound, Shield, CheckCircle, AlertTriangle, User } from "lucide-react";

export default function SettingsPage() {
  const router = useRouter();
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [userEmail, setUserEmail] = useState("admin@vulnex.local");

  useEffect(() => {
    if (typeof window !== "undefined") {
      const token = localStorage.getItem("vulnex_token");
      if (!token) {
        router.push("/login");
        return;
      }
      const stored = localStorage.getItem("vulnex_user");
      if (stored) {
        try {
          const parsed = JSON.parse(stored);
          if (parsed.email) setUserEmail(parsed.email);
        } catch {}
      }
    }
  }, [router]);

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    if (!currentPassword) {
      setError("Current password is required.");
      return;
    }

    if (newPassword !== confirmPassword) {
      setError("New passwords do not match.");
      return;
    }

    if (newPassword.length < 5) {
      setError("New password must be at least 5 characters long.");
      return;
    }

    setLoading(true);

    try {
      await api.changePassword(currentPassword, newPassword);
      setSuccess("Password successfully changed!");
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    } catch (err: any) {
      setError(err.message || "Failed to change password.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-white flex items-center space-x-2">
          <KeyRound className="w-6 h-6 text-blue-400" />
          <span>Account Settings & Security</span>
        </h1>
        <p className="text-sm text-slate-400 mt-1">
          Manage your operator credentials, password, and security preferences.
        </p>
      </div>

      <div className="p-6 rounded-xl bg-vulnex-card border border-vulnex-border space-y-6">
        <div className="flex items-center space-x-4 pb-6 border-b border-vulnex-border">
          <div className="p-3 rounded-full bg-blue-600/20 text-blue-400 border border-blue-500/30">
            <User className="w-8 h-8" />
          </div>
          <div>
            <h2 className="text-lg font-bold text-white">Administrator Profile</h2>
            <div className="font-mono text-xs text-slate-400">{userEmail}</div>
            <div className="inline-block mt-1 px-2 py-0.5 rounded text-[10px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              ACTIVE OPERATOR
            </div>
          </div>
        </div>

        <div>
          <h3 className="text-base font-semibold text-white mb-2 flex items-center space-x-2">
            <Shield className="w-4 h-4 text-blue-400" />
            <span>Change Operator Password</span>
          </h3>
          <p className="text-xs text-slate-400 mb-4">
            Update default credentials to protect your VULNEX deployment.
          </p>

          {error && (
            <div className="mb-4 p-3 rounded-lg bg-red-500/10 border border-red-500/30 text-red-400 text-xs flex items-start space-x-2">
              <AlertTriangle className="w-4 h-4 flex-shrink-0 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          {success && (
            <div className="mb-4 p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-xs flex items-start space-x-2">
              <CheckCircle className="w-4 h-4 flex-shrink-0 mt-0.5" />
              <span>{success}</span>
            </div>
          )}

          <form onSubmit={handleChangePassword} className="space-y-4">
            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1">
                Current Password (default: admin)
              </label>
              <input
                type="password"
                required
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                placeholder="••••••••"
                className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-lg text-sm text-white focus:outline-none focus:border-blue-500"
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">
                  New Password
                </label>
                <input
                  type="password"
                  required
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  placeholder="Enter new password"
                  className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-lg text-sm text-white focus:outline-none focus:border-blue-500"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">
                  Confirm New Password
                </label>
                <input
                  type="password"
                  required
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  placeholder="Re-enter new password"
                  className="w-full px-3 py-2 bg-slate-900 border border-slate-700 rounded-lg text-sm text-white focus:outline-none focus:border-blue-500"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="py-2.5 px-6 bg-blue-600 hover:bg-blue-500 disabled:bg-slate-700 text-white font-medium rounded-lg text-sm transition shadow flex items-center space-x-2"
            >
              <span>{loading ? "Updating Password..." : "Update Password"}</span>
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
