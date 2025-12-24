"use client";

import { useState, useEffect } from "react";
import { Copy, Check } from "lucide-react";
import { useTheme } from "../theme/ThemeProvider";
import { codeToHtml } from "shiki";

interface CodeBlockProps {
  children: string;
  className?: string;
  language?: string;
}

export function CodeBlock({ children, className, language }: CodeBlockProps) {
  const { theme, isDark } = useTheme();
  const [copied, setCopied] = useState(false);
  const [highlightedCode, setHighlightedCode] = useState<string>("");
  const [detectedLanguage, setDetectedLanguage] = useState<string>("");

  // Extract language from className (format: language-xxx)
  useEffect(() => {
    const langFromClass = className?.match(/language-(\w+)/)?.[1] || language || "text";
    setDetectedLanguage(langFromClass);
  }, [className, language]);

  // Highlight code with Shiki
  useEffect(() => {
    const highlight = async () => {
      try {
        const html = await codeToHtml(children, {
          lang: detectedLanguage === "text" ? "plaintext" : detectedLanguage,
          theme: isDark ? "github-dark" : "github-light",
        });
        // Remove background color from Shiki output to match container
        const modifiedHtml = html.replace(
          /background-color:\s*[^;]+;?/gi,
          'background-color: transparent;'
        );
        setHighlightedCode(modifiedHtml);
      } catch (error) {
        // Fallback to plain text if highlighting fails
        setHighlightedCode(
          `<pre><code>${children.replace(/</g, "&lt;").replace(/>/g, "&gt;")}</code></pre>`
        );
      }
    };
    highlight();
  }, [children, detectedLanguage, isDark]);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(children);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error("Failed to copy:", err);
    }
  };

  return (
    <div
      className={`${theme.cardBg} border ${theme.border} rounded-lg overflow-hidden relative group transition-colors duration-300`}
    >
      {/* Language label and copy button */}
      <div className={`flex items-center justify-between px-4 py-2 border-b ${theme.border}`}>
        <span className={`text-xs ${theme.subtext} uppercase tracking-widest font-mono`}>
          {detectedLanguage}
        </span>
        <button
          onClick={handleCopy}
          className={`flex items-center gap-2 text-xs ${theme.subtext} hover:${theme.accent} transition-colors`}
          aria-label="Copy code"
        >
          {copied ? (
            <>
              <Check size={14} />
              <span>Copied</span>
            </>
          ) : (
            <>
              <Copy size={14} />
              <span>Copy</span>
            </>
          )}
        </button>
      </div>

      {/* Code content */}
      <div className={`overflow-x-auto ${theme.cardBg}`}>
        <div
          className="p-4 text-sm [&_pre]:!m-0 [&_pre]:!p-0"
          dangerouslySetInnerHTML={{ __html: highlightedCode }}
        />
      </div>
    </div>
  );
}

