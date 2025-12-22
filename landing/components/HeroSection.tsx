"use client";

import { UrlShortener } from "./UrlShortener";
import { CodeSnippet } from "./CodeSnippet";
import { useTheme } from "./theme/ThemeProvider";

export function HeroSection() {
  const { theme, isDark } = useTheme();

  return (
    <section className="space-y-8">
      <div className="space-y-4">
        <p
          className={`text-lg leading-relaxed ${
            isDark ? "text-gray-200" : "text-gray-800"
          }`}
        >
          Shorten links via CLI, API, or Web. <br />
          MIT Licensed. Self-hostable.{" "}
          <span className={theme.accent}> Bloat-free.</span>
        </p>
      </div>

      {/* Interactive Shortener UI */}
      <UrlShortener />

      {/* Code Snippet */}
      <CodeSnippet />
    </section>
  );
}

