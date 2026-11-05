import type { Metadata } from "next";
import Link from "next/link";
import "./globals.css";
import { Shield, LayoutDashboard, Globe, AlertTriangle, Layers, Settings, LogOut } from "lucide-react";
import NavbarUser from "@/components/NavbarUser";

export const metadata: Metadata = {
  title: "VULNEX — Vulnerability Assessment & Pen-Testing Management",
  description: "Defensive vulnerability assessment and penetration testing management platform",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="dark">
      <body className="flex flex-col min-h-screen bg-vulnex-bg text-slate-100 font-sans antialiased">
        <header className="border-b border-vulnex-border bg-vulnex-card/80 backdrop-blur sticky top-0 z-50">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
            <div className="flex items-center space-x-8">
              <Link href="/" className="flex items-center space-x-3 group">
                <div className="p-2 rounded-lg bg-blue-600/20 text-blue-400 border border-blue-500/30 group-hover:border-blue-400 transition">
                  <Shield className="w-6 h-6" />
                </div>
                <div className="flex flex-col">
                  <span className="font-bold tracking-wider text-lg text-white">VULNEX</span>
                  <span className="text-[10px] text-slate-400 font-mono -mt-1 tracking-widest">DEFENSIVE PLATFORM</span>
                </div>
              </Link>

              <nav className="hidden md:flex space-x-1">
                <Link
                  href="/"
                  className="px-3 py-2 rounded-md text-sm font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition flex items-center space-x-2"
                >
                  <LayoutDashboard className="w-4 h-4" />
                  <span>Dashboard</span>
                </Link>
                <Link
                  href="/workspaces"
                  className="px-3 py-2 rounded-md text-sm font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition flex items-center space-x-2"
                >
                  <Layers className="w-4 h-4" />
                  <span>Workspaces</span>
                </Link>
                <Link
                  href="/scans"
                  className="px-3 py-2 rounded-md text-sm font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition flex items-center space-x-2"
                >
                  <Globe className="w-4 h-4" />
                  <span>Scans</span>
                </Link>
                <Link
                  href="/findings"
                  className="px-3 py-2 rounded-md text-sm font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition flex items-center space-x-2"
                >
                  <AlertTriangle className="w-4 h-4" />
                  <span>Findings</span>
                </Link>
                <Link
                  href="/settings"
                  className="px-3 py-2 rounded-md text-sm font-medium text-slate-300 hover:text-white hover:bg-slate-800/60 transition flex items-center space-x-2"
                >
                  <Settings className="w-4 h-4" />
                  <span>Settings</span>
                </Link>
              </nav>
            </div>

            <NavbarUser />
          </div>
        </header>

        <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {children}
        </main>

        <footer className="border-t border-vulnex-border py-4 text-center text-xs text-slate-500">
          VULNEX Platform v1.0.0 • Defensive Pen-Testing Management • Apache 2.0
        </footer>
      </body>
    </html>
  );
}
