"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";
import { User, LogOut, Settings, LogIn } from "lucide-react";

interface StoredUser {
  id?: string;
  email?: string;
  display_name?: string;
}

export default function NavbarUser() {
  const router = useRouter();
  const [token, setToken] = useState<string | null>(null);
  const [user, setUser] = useState<StoredUser | null>(null);

  const checkAuth = () => {
    if (typeof window !== "undefined") {
      const storedToken = localStorage.getItem("vulnex_token");
      const storedUserRaw = localStorage.getItem("vulnex_user");
      setToken(storedToken);
      if (storedUserRaw) {
        try {
          setUser(JSON.parse(storedUserRaw));
        } catch {
          setUser(null);
        }
      } else {
        setUser(null);
      }
    }
  };

  useEffect(() => {
    checkAuth();
    const handleAuthChange = () => checkAuth();
    window.addEventListener("vulnex:auth-change", handleAuthChange);
    window.addEventListener("storage", handleAuthChange);
    return () => {
      window.removeEventListener("vulnex:auth-change", handleAuthChange);
      window.removeEventListener("storage", handleAuthChange);
    };
  }, []);

  const handleSignOut = () => {
    api.clearToken();
    setToken(null);
    setUser(null);
    router.push("/login");
  };

  if (token) {
    const displayName = user?.display_name || user?.email?.split("@")[0] || "admin";
    return (
      <div className="flex items-center space-x-3">
        <Link
          href="/settings"
          className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-slate-800/80 hover:bg-slate-700/80 border border-slate-700 text-xs font-medium text-slate-200 transition"
          title="Account Settings & Password"
        >
          <User className="w-3.5 h-3.5 text-blue-400" />
          <span className="font-semibold text-white">{displayName}</span>
        </Link>
        <Link
          href="/settings"
          className="p-1.5 rounded-lg bg-slate-800/80 hover:bg-slate-700 border border-slate-700 text-slate-400 hover:text-white transition"
          title="Change Password"
        >
          <Settings className="w-4 h-4" />
        </Link>
        <button
          onClick={handleSignOut}
          className="flex items-center space-x-1 text-xs bg-red-600/20 hover:bg-red-600/30 text-red-400 hover:text-red-300 border border-red-500/30 px-3 py-1.5 rounded transition"
          title="Sign Out"
        >
          <LogOut className="w-3.5 h-3.5" />
          <span>Sign Out</span>
        </button>
      </div>
    );
  }

  return (
    <div className="flex items-center space-x-3">
      <Link
        href="/setup"
        className="text-xs bg-slate-800 text-slate-300 hover:text-white px-3 py-1.5 rounded border border-slate-700 transition"
      >
        Onboarding Setup
      </Link>
      <Link
        href="/login"
        className="text-xs bg-blue-600 hover:bg-blue-500 text-white font-medium px-3 py-1.5 rounded transition shadow-sm flex items-center space-x-1.5"
      >
        <LogIn className="w-3.5 h-3.5" />
        <span>Sign In</span>
      </Link>
    </div>
  );
}
