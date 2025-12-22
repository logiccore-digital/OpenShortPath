'use client'

import { useTheme } from './theme/ThemeProvider'

export function Footer() {
  const { theme } = useTheme()

  return (
    <footer
      className={`pt-16 pb-8 border-t ${theme.border} flex flex-col sm:flex-row justify-between items-center gap-4 text-xs ${theme.subtext} transition-colors duration-300`}
    >
      <div className="flex gap-6">
        <a href="#" className="hover:underline">
          GitHub
        </a>
        <a href="#" className="hover:underline">
          Docs
        </a>
        <a href="#" className="hover:underline">
          Terms
        </a>
        <a href="#" className="hover:underline">
          Privacy
        </a>
      </div>
      <div>© {new Date().getFullYear()} OpenShortPath.</div>
    </footer>
  )
}

