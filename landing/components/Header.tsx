'use client'

import Link from 'next/link'
import { Terminal, Github } from 'lucide-react'
import { useTheme } from './theme/ThemeProvider'

export function Header() {
  const { isDark, toggleTheme, theme } = useTheme()

  return (
    <header className={`flex justify-between items-start border-b ${theme.border} pb-8 transition-colors duration-300`}>
      <div className="space-y-2">
        <Link href="/">
          <h1 className={`text-2xl font-bold ${theme.heading} tracking-tighter flex items-center gap-2 hover:opacity-80 transition-opacity cursor-pointer`}>
            <Terminal size={20} className={theme.accent} />
            OpenShortPath_
          </h1>
        </Link>
        <p className={`${theme.subtext} text-sm max-w-sm`}>
          The no-nonsense, open-source link shortener for developers.
        </p>
      </div>
      <div className="flex flex-col items-end gap-3">
        <button
          onClick={toggleTheme}
          className={`${theme.subtext} hover:${theme.heading} text-xs uppercase tracking-wider flex items-center gap-2 transition-colors`}
        >
          [{isDark ? 'light_mode' : 'dark_mode'}]
        </button>
        <a
          href="#"
          className={`group flex items-center gap-2 text-sm ${theme.subtext} hover:${theme.heading} transition-colors`}
        >
          <Github size={16} />
          <span className="hidden sm:inline group-hover:underline decoration-1 underline-offset-4">/openshortpath</span>
        </a>
      </div>
    </header>
  )
}

