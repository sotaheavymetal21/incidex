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

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ja">
      <body className="antialiased bg-gray-50">
        <AuthProvider>
          <Navbar />
          <main>
            {children}
          </main>
        </AuthProvider>
      </body>
    </html>
  );
}
