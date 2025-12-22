import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { ThemeProvider } from "@/components/theme/ThemeProvider";
import type { ReactNode } from "react";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "OpenShortPath - The no-nonsense, open-source link shortener",
  description: "The no-nonsense, open-source link shortener for developers.",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={`${inter.className} font-mono`}>
        <ThemeProvider>{children}</ThemeProvider>
      </body>
    </html>
  );
}
