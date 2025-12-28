import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Incidex - インシデント管理システム",
  description: "モダンなインシデント管理システム",
  icons: {
    icon: "/incidex_i_logo.jpg",
  },
};

import { AuthProvider } from "@/context/AuthContext";
import Navbar from "@/components/Navbar";
import { ErrorBoundary } from "@/components/ErrorBoundary";

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ja">
      <body className="antialiased" style={{ background: 'var(--background)', color: 'var(--foreground)' }}>
        <ErrorBoundary>
          <AuthProvider>
            <Navbar />
            <main>
              {children}
            </main>
          </AuthProvider>
        </ErrorBoundary>
      </body>
    </html>
  );
}
