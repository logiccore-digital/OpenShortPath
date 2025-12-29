'use client'

import Link from 'next/link'
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
        <Link href="/docs" className="hover:underline">
          Docs
        </Link>
        <Link href="/docs/terms-of-service" className="hover:underline">
          Terms
        </Link>
        <Link href="/docs/privacy-policy" className="hover:underline">
          Privacy
        </Link>
      </div>
      <div>© {new Date().getFullYear()} OpenShortPath.</div>
    </footer>
  )
}

