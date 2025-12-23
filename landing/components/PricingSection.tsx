'use client'

import { Check, Command, Shield } from 'lucide-react'
import { useTheme } from './theme/ThemeProvider'

export function PricingSection() {
  const { isDark, theme } = useTheme()

  return (
    <section className={`space-y-8 border-t ${theme.border} pt-16 transition-colors duration-300`}>
      <div className="space-y-2">
        <h2 className={`text-xl font-bold ${theme.heading}`}>Hosted Usage & Limits</h2>
        <p className={`${theme.subtext} text-sm`}>
          We charge for API access to prevent abuse and keep the service fast for everyone.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-12 border-b border-gray-800/50 pb-8">
        {/* Free Tier */}
        <div className="space-y-4">
          <div className={`h-1 w-8 ${isDark ? 'bg-gray-700' : 'bg-gray-300'}`}></div>
          <h3 className={`font-bold ${theme.heading}`}>Hobbyist</h3>
          <div className={`text-2xl font-bold ${theme.heading}`}>
            $0<span className={`text-sm font-normal ${theme.subtext}`}>/mo</span>
          </div>
          <ul className={`text-sm space-y-3 ${theme.subtext}`}>
            <li className="flex gap-2">
              <Check size={14} className="opacity-70 shrink-0 mt-0.5" /> 1,000 links / month
            </li>
            <li className="flex gap-2">
              <Check size={14} className="opacity-70 shrink-0 mt-0.5" /> Web Interface
            </li>
            <li className="flex gap-2">
              <Check size={14} className="opacity-70 shrink-0 mt-0.5" /> 1 Namespace
            </li>
            <li className="flex gap-2">
              <Command size={14} className="opacity-70 shrink-0 mt-0.5" /> Limited API/CLI Access (5 per hour)
            </li>
            <li className="flex gap-2 opacity-50">
              <Command size={14} className="opacity-70 shrink-0 mt-0.5" /> No Webhooks
            </li>
          </ul>
        </div>

        {/* Pro Plan */}
        <div className="space-y-4">
          <div className="h-1 w-8 bg-purple-500"></div>
          <h3 className={`font-bold ${theme.heading}`}>Pro</h3>
          <div className={`text-2xl font-bold ${theme.heading}`}>
            $5<span className={`text-sm font-normal ${theme.subtext}`}>/mo</span>
          </div>
          <ul className={`text-sm space-y-3 ${theme.subtext}`}>
            <li className="flex gap-2">
              <Check size={14} className="text-purple-500 shrink-0 mt-0.5" />{' '}
              <strong>10,000 links / month</strong>
            </li>
            <li className="flex gap-2">
              <Check size={14} className="text-purple-500 shrink-0 mt-0.5" /> Full API & CLI Access
            </li>
            <li className="flex gap-2">
              <Check size={14} className="text-purple-500 shrink-0 mt-0.5" /> 10 Namespaces
            </li>
            <li className="flex gap-2">
              <Check size={14} className="text-purple-500 shrink-0 mt-0.5" /> Webhooks & Events
            </li>
            <li className="flex gap-2">
              <Check size={14} className="text-purple-500 shrink-0 mt-0.5" /> Priority Support
            </li>
          </ul>
          <button
            className={`w-full py-2 border ${theme.border} ${isDark ? 'text-gray-300' : 'text-gray-600'} text-sm hover:${theme.heading} ${theme.hoverBorder} transition-colors mt-2`}
          >
            Subscribe
          </button>
        </div>
      </div>

      {/* Verified Option - Separate Section */}
      <div className="space-y-4">
        <div className="flex items-center gap-3">
          <Shield size={16} className="text-emerald-500" />
          <h3 className={`font-bold ${theme.heading}`}>Verified Access</h3>
        </div>
        <div className={`flex flex-col md:flex-row justify-between gap-6 border ${theme.border} p-6 ${theme.cardBg}`}>
          <div className="space-y-3">
            <p className={`text-sm ${theme.subtext} max-w-md`}>
              Don&apos;t want a subscription? Pay a one-time fee to verify your identity. This unlocks full API/CLI
              access with the standard 1,000 links/month limit and removes the 5 per hour rate limit.
            </p>
            <ul className={`text-sm space-y-1 ${theme.subtext}`}>
              <li className="flex gap-2 items-center">
                <Check size={12} className="text-emerald-500" /> Lifetime API Access
              </li>
              <li className="flex gap-2 items-center">
                <Check size={12} className="text-emerald-500" /> No 5 per hour rate limit
              </li>
              <li className="flex gap-2 items-center">
                <Check size={12} className="text-emerald-500" /> No recurring fees
              </li>
            </ul>
          </div>
          <div className="flex flex-col items-start md:items-end gap-3 min-w-[140px]">
            <div className={`text-2xl font-bold ${theme.heading}`}>
              $10<span className={`text-sm font-normal ${theme.subtext}`}> one-time</span>
            </div>
            <button className="px-6 py-2 border border-emerald-800 text-emerald-500 text-sm hover:bg-emerald-900/20 transition-colors w-full md:w-auto">
              Verify Identity
            </button>
          </div>
        </div>
      </div>

      <div className="pt-4">
        <div className={`border border-dashed ${theme.border} p-4 flex flex-col sm:flex-row justify-between items-center gap-4`}>
          <div className={`text-sm ${theme.subtext}`}>
            <span className={`${theme.heading} block mb-1`}>Building a SaaS?</span>
            Need higher limits for platform integration? We offer custom volume pricing.
          </div>
          <a
            href="mailto:hello@openshortpath.com"
            className={`whitespace-nowrap px-4 py-2 ${
              isDark ? 'bg-gray-100 text-black hover:bg-white' : 'bg-gray-900 text-white hover:bg-gray-700'
            } text-sm font-bold transition-colors`}
          >
            Contact Sales
          </a>
        </div>
      </div>
    </section>
  )
}

