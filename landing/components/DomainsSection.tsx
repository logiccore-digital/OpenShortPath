'use client'

import { Hash, Heart } from 'lucide-react'
import { useTheme } from './theme/ThemeProvider'

export function DomainsSection() {
  const { theme } = useTheme()

  return (
    <section className="grid grid-cols-1 md:grid-cols-2 gap-6 pt-4">
      <div className={`border ${theme.border} ${theme.hoverBorder} p-5 transition-colors duration-300`}>
        <h3 className={`${theme.heading} font-bold mb-2 flex items-center gap-2`}>
          <Hash size={16} className="text-gray-500" /> Hosted Domains
        </h3>
        <p className={`text-xs ${theme.subtext} mb-4`}>
          Choose your preferred short domain. More domains will be added to the hosted version soon.
        </p>
        <ul className={`space-y-2 text-sm ${theme.subtext}`}>
          <li className="flex justify-between items-center">
            <span>lcd.sh</span>
            <span className="text-emerald-500 text-[10px] uppercase tracking-wider border border-emerald-900 bg-emerald-900/10 px-1.5 py-0.5 rounded">
              Available
            </span>
          </li>
          <li className="flex justify-between items-center">
            <span>mix.lol</span>
            <span className="text-emerald-500 text-[10px] uppercase tracking-wider border border-emerald-900 bg-emerald-900/10 px-1.5 py-0.5 rounded">
              Available
            </span>
          </li>
        </ul>
      </div>
      <div className={`border ${theme.border} ${theme.hoverBorder} p-5 transition-colors duration-300`}>
        <h3 className={`${theme.heading} font-bold mb-2 flex items-center gap-2`}>
          <Heart size={16} className="text-gray-500" /> Open Source
        </h3>
        <p className={`text-sm ${theme.subtext} leading-relaxed`}>
          Check the code, fork it, or host it yourself. We believe in transparent infrastructure. Released under MIT
          License.
        </p>
      </div>
    </section>
  )
}

