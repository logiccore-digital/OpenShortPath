"use client";

import { useState } from "react";
import { Copy, Check, ChevronRight } from "lucide-react";
import { useTheme } from "./theme/ThemeProvider";

export function CodeSnippet() {
  const { theme } = useTheme();
  const [copied, setCopied] = useState(false);
  const [activeTab, setActiveTab] = useState<"curl" | "js">("curl");
  const [selectedDomain] = useState("lcd.sh"); // This could be passed as prop if needed

  const handleCopy = () => {
    let codeToCopy = "";
    if (activeTab === "curl") {
      codeToCopy = `curl -X POST https://api.lcd.sh/new \\
  -H "Authorization: Bearer <token>" \\
  -d "url=https://github.com/openshortpath" \\
  -d "domain=${selectedDomain}"`;
    } else {
      codeToCopy = `const res = await fetch('https://api.lcd.sh/new', {
  method: 'POST',
  headers: { 'Authorization': 'Bearer <token>' },
  body: JSON.stringify({
    url: '...',
    domain: '${selectedDomain}'
  })
});`;
    }
    navigator.clipboard.writeText(codeToCopy);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div>
      <div
        className={`flex items-center gap-2 mb-3 text-xs ${theme.subtext} uppercase tracking-widest`}
      >
        <ChevronRight size={14} /> API Usage
      </div>
      <div
        className={`${theme.cardBg} border ${theme.border} rounded p-4 relative group transition-colors duration-300`}
      >
        <div className="absolute top-4 right-4 flex gap-2">
          <div
            className={`text-xs ${theme.subtext} uppercase tracking-widest pt-1`}
          >
            bash
          </div>
        </div>
        <div className={`flex gap-4 mb-4 border-b ${theme.border} pb-2`}>
          <button
            onClick={() => setActiveTab("curl")}
            className={`text-sm pb-2 -mb-2.5 transition-colors ${
              activeTab === "curl"
                ? `${theme.accent} border-b ${theme.accentBorder}`
                : `${theme.subtext} hover:${theme.heading}`
            }`}
          >
            cURL
          </button>
          <button
            onClick={() => setActiveTab("js")}
            className={`text-sm pb-2 -mb-2.5 transition-colors ${
              activeTab === "js"
                ? `${theme.accent} border-b ${theme.accentBorder}`
                : `${theme.subtext} hover:${theme.heading}`
            }`}
          >
            Node.js
          </button>
        </div>

        <div className="overflow-x-auto">
          {activeTab === "curl" ? (
            <pre className="text-sm whitespace-pre">
              <code>
                <span className={theme.code.keyword}>curl</span> -X POST
                https://api.lcd.sh/new \<br />
                &nbsp;&nbsp;-H{" "}
                <span className={theme.code.attr}>
                  &quot;Authorization: Bearer &lt;token&gt;&quot;
                </span>{" "}
                \<br />
                &nbsp;&nbsp;-d{" "}
                <span className={theme.code.string}>
                  &quot;url=https://github.com/openshortpath&quot;
                </span>{" "}
                \<br />
                &nbsp;&nbsp;-d{" "}
                <span className={theme.code.string}>
                  &quot;domain={selectedDomain}&quot;
                </span>
              </code>
            </pre>
          ) : (
            <pre className="text-sm whitespace-pre">
              <code>
                <span className={theme.code.keyword}>const</span> res ={" "}
                <span className={theme.code.keyword}>await</span> fetch(
                <span className={theme.code.string}>
                  &apos;https://api.lcd.sh/new&apos;
                </span>
                , {"{"}
                <br />
                &nbsp;&nbsp;method:{" "}
                <span className={theme.code.string}>&apos;POST&apos;</span>,
                <br />
                &nbsp;&nbsp;headers: {"{"}{" "}
                <span className={theme.code.string}>
                  &apos;Authorization&apos;
                </span>
                :{" "}
                <span className={theme.code.string}>
                  &apos;Bearer &lt;token&gt;&apos;
                </span>{" "}
                {"}"},<br />
                &nbsp;&nbsp;body: JSON.stringify({"{"}
                <br />
                &nbsp;&nbsp;&nbsp;&nbsp;url:{" "}
                <span className={theme.code.string}>&apos;...&apos;</span>,
                <br />
                &nbsp;&nbsp;&nbsp;&nbsp;domain:{" "}
                <span className={theme.code.string}>
                  &apos;{selectedDomain}&apos;
                </span>
                <br />
                &nbsp;&nbsp;{"}"})<br />
                {"}"});
              </code>
            </pre>
          )}
        </div>

        <button
          onClick={handleCopy}
          className={`mt-4 text-xs flex items-center gap-2 ${theme.subtext} hover:${theme.accent} transition-colors`}
        >
          {copied ? <Check size={12} /> : <Copy size={12} />}
          {copied ? "Copied to clipboard" : "Copy snippet"}
        </button>
      </div>
    </div>
  );
}
