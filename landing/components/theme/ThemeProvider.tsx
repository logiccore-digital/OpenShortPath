"use client";

import React, { createContext, useContext, useState, useEffect } from "react";

interface ThemeContextType {
  isDark: boolean;
  toggleTheme: () => void;
  theme: {
    bg: string;
    text: string;
    heading: string;
    border: string;
    hoverBorder: string;
    subtext: string;
    cardBg: string;
    inputBg: string;
    code: {
      keyword: string;
      string: string;
      attr: string;
    };
    selection: string;
    accent: string;
    accentBorder: string;
    buttonText: string;
  };
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [isDark, setIsDark] = useState(true);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    // Check localStorage for saved theme preference
    const saved = localStorage.getItem("theme");
    if (saved === "light") {
      setIsDark(false);
    }
  }, []);

  useEffect(() => {
    if (mounted) {
      if (isDark) {
        document.documentElement.classList.add("dark");
        localStorage.setItem("theme", "dark");
      } else {
        document.documentElement.classList.remove("dark");
        localStorage.setItem("theme", "light");
      }
    }
  }, [isDark, mounted]);

  const toggleTheme = () => {
    setIsDark(!isDark);
  };

  const theme = {
    bg: isDark ? "bg-[#0d0d0d]" : "bg-white",
    text: isDark ? "text-gray-300" : "text-gray-600",
    heading: isDark ? "text-gray-100" : "text-gray-900",
    border: isDark ? "border-gray-800" : "border-gray-200",
    hoverBorder: isDark ? "hover:border-gray-600" : "hover:border-gray-400",
    subtext: "text-gray-500",
    cardBg: isDark ? "bg-[#111]" : "bg-gray-50",
    inputBg: isDark ? "bg-[#0a0a0a]" : "bg-white",
    code: {
      keyword: isDark ? "text-purple-400" : "text-purple-600",
      string: isDark ? "text-yellow-300" : "text-amber-600",
      attr: isDark ? "text-green-400" : "text-emerald-600",
    },
    selection: isDark
      ? "selection:bg-gray-700 selection:text-white"
      : "selection:bg-gray-200 selection:text-black",
    accent: "text-emerald-500",
    accentBorder: "border-emerald-500",
    buttonText: isDark ? "text-black" : "text-white",
  };

  // Prevent hydration mismatch
  if (!mounted) {
    return <div className="min-h-screen bg-[#0d0d0d]">{children}</div>;
  }

  return (
    <ThemeContext.Provider value={{ isDark, toggleTheme, theme }}>
      <div
        className={`min-h-screen ${theme.bg} ${theme.text} ${theme.selection} transition-colors duration-300`}
      >
        {children}
      </div>
    </ThemeContext.Provider>
  );
}

export function useTheme() {
  const context = useContext(ThemeContext);
  if (context === undefined) {
    // Provide default theme during static generation
    const defaultIsDark = true;
    const defaultTheme = {
      bg: "bg-[#0d0d0d]",
      text: "text-gray-300",
      heading: "text-gray-100",
      border: "border-gray-800",
      hoverBorder: "hover:border-gray-600",
      subtext: "text-gray-500",
      cardBg: "bg-[#111]",
      inputBg: "bg-[#0a0a0a]",
      code: {
        keyword: "text-purple-400",
        string: "text-yellow-300",
        attr: "text-green-400",
      },
      selection: "selection:bg-gray-700 selection:text-white",
      accent: "text-emerald-500",
      accentBorder: "border-emerald-500",
      buttonText: "text-black",
    };
    return {
      isDark: defaultIsDark,
      toggleTheme: () => {},
      theme: defaultTheme,
    };
  }
  return context;
}
