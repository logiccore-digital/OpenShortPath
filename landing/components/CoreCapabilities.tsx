"use client";

import { Zap, ArrowLeftRight, FolderTree, Webhook, Edit3 } from "lucide-react";
import { useTheme } from "./theme/ThemeProvider";

export function CoreCapabilities() {
  const { isDark, theme } = useTheme();

  const features = [
    {
      icon: ArrowLeftRight,
      title: "301 Redirects",
      description:
        "Standard, simple HTTP 301 Permanent redirects. Optimized for speed and ensures zero SEO penalty for your links.",
    },
    {
      icon: FolderTree,
      title: "Namespaces",
      description: (
        <>
          Organize paths with custom namespaces. Support for{" "}
          <code
            className={`font-mono text-xs ${
              isDark ? "bg-gray-800/50" : "bg-gray-100"
            } ${theme.heading} px-1 py-0.5 rounded`}
          >
            domain.com/{`<`}namespace{`>`}/{`<`}slug{`>`}
          </code>{" "}
          patterns for team isolation.
        </>
      ),
    },
    {
      icon: Webhook,
      title: "Visit Webhooks",
      description:
        "Event-driven architecture. Configure webhooks to trigger external automation or analytics whenever a link is visited.",
    },
    {
      icon: Edit3,
      title: "Custom Slugs",
      description:
        "Don't like random strings? Specify your own custom aliases (e.g., /launch-day) to make your links memorable and on-brand.",
    },
  ];

  return (
    <section className="pt-8">
      <h2
        className={`${theme.heading} text-xl font-bold mb-6 flex items-center gap-2`}
      >
        <Zap size={20} className={theme.accent} />
        Core Capabilities
      </h2>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {features.map((feature) => (
          <div
            key={feature.title}
            className={`p-5 border ${theme.border} ${theme.cardBg} flex flex-col items-start transition-colors duration-300`}
          >
            <div
              className={`mb-3 p-2 rounded ${
                isDark ? "bg-gray-800/30" : "bg-gray-100"
              } text-emerald-500`}
            >
              <feature.icon size={20} />
            </div>
            <h3 className={`font-bold ${theme.heading} mb-2`}>
              {feature.title}
            </h3>
            <p className={`text-sm ${theme.subtext} leading-relaxed`}>
              {typeof feature.description === "string"
                ? feature.description
                : feature.description}
            </p>
          </div>
        ))}
      </div>
    </section>
  );
}
