'use client'

import { Server, Database, Box, Zap, Layers } from 'lucide-react'
import { useTheme } from './theme/ThemeProvider'

export function SelfHostingSection() {
  const { isDark, theme } = useTheme()

  const features = [
    {
      icon: Database,
      title: 'SQLite by default.',
      description: 'No external dependencies. Just run the binary and it works.',
    },
    {
      icon: Box,
      title: 'Docker Support.',
      description: 'Official images available. Zero config required to get started.',
    },
    {
      icon: Zap,
      title: 'In-memory cache.',
      description: 'Built-in high performance caching without needing Redis.',
    },
    {
      icon: Layers,
      title: 'Scale when ready.',
      description: 'Need to scale? Simply configure Postgres and Redis in config.yaml.',
    },
  ]

  return (
    <section className="grid grid-cols-1 md:grid-cols-2 gap-12 pt-8 items-center">
      <div className="space-y-6">
        <div className="space-y-2">
          <h2 className={`${theme.heading} text-xl font-bold flex items-center gap-2`}>
            <Server size={20} className={theme.accent} />
            Easy Self-Hosting
          </h2>
          <p className={`${theme.subtext} leading-relaxed`}>
            Deployment shouldn&apos;t be a headache. OpenShortPath is compiled to a single <strong>Go binary</strong>.
          </p>
        </div>

        <ul className={`space-y-4 text-sm ${theme.subtext}`}>
          {features.map((feature) => (
            <li key={feature.title} className="flex gap-3">
              <div
                className={`mt-0.5 p-1 rounded ${isDark ? 'bg-gray-800/50' : 'bg-gray-100'} h-fit`}
              >
                <feature.icon size={14} className={theme.heading} />
              </div>
              <div>
                <strong className={theme.heading}>{feature.title}</strong> <br />
                {feature.description}
              </div>
            </li>
          ))}
        </ul>
      </div>

      <div>
        <div className={`${theme.cardBg} border ${theme.border} rounded overflow-hidden`}>
          <div
            className={`text-xs ${theme.subtext} uppercase tracking-widest p-3 border-b ${theme.border} bg-opacity-50 flex justify-between`}
          >
            <span>Terminal</span>
            <span>~</span>
          </div>
          <div className="p-4 overflow-x-auto">
            <code className="text-sm block whitespace-pre font-mono">
              <span className={theme.subtext}># It&apos;s really this simple</span>
              <br />
              <span className="text-emerald-500 font-bold">$</span> curl -L -o openshortpath https://...
              <br />
              <span className="text-emerald-500 font-bold">$</span> ./openshortpath start
              <br />
              <br />
              <span className={theme.subtext}># Or use Docker</span>
              <br />
              <span className="text-emerald-500 font-bold">$</span> docker run -p 8080:8080 openshortpath/server
              <br />
              <br />
              <span className={`block mt-3 ${theme.subtext} opacity-70`}>
                [INFO] Starting server on :8080<br />
                [INFO] Storage: SQLite (data.db)<br />
                [INFO] Cache: In-memory (LRU)<br />
                [INFO] Ready to shorten!
              </span>
            </code>
          </div>
        </div>
      </div>
    </section>
  )
}

