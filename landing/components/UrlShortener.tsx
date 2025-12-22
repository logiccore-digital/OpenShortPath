'use client'

import { useState } from 'react'
import { ArrowRight, Loader2, Copy, Check, ChevronDown } from 'lucide-react'
import { useTheme } from './theme/ThemeProvider'

export function UrlShortener() {
  const { theme, isDark } = useTheme()
  const [inputUrl, setInputUrl] = useState('')
  const [selectedDomain, setSelectedDomain] = useState('lcd.sh')
  const [loading, setLoading] = useState(false)
  const [shortenedResult, setShortenedResult] = useState('')
  const [resultCopied, setResultCopied] = useState(false)

  const handleShorten = (e: React.FormEvent) => {
    e.preventDefault()
    if (!inputUrl) return

    setLoading(true)
    // Simulate API call
    setTimeout(() => {
      setShortenedResult(`https://${selectedDomain}/${Math.random().toString(36).substr(2, 6)}`)
      setLoading(false)
    }, 800)
  }

  const handleResultCopy = () => {
    navigator.clipboard.writeText(shortenedResult)
    setResultCopied(true)
    setTimeout(() => setResultCopied(false), 2000)
  }

  return (
    <div className="space-y-2">
      <div className="flex justify-between items-end">
        <label className={`text-xs uppercase tracking-widest ${theme.subtext}`}>Try it now (Free)</label>
        {shortenedResult && (
          <button
            onClick={() => {
              setShortenedResult('')
              setInputUrl('')
            }}
            className={`text-xs ${theme.subtext} hover:${theme.heading} underline`}
          >
            Reset
          </button>
        )}
      </div>

      {!shortenedResult ? (
        <form onSubmit={handleShorten} className="flex flex-col sm:flex-row gap-0 relative">
          {/* Domain Selector */}
          <div
            className={`relative ${theme.inputBg} border ${theme.border} sm:border-r-0 flex items-center min-w-[140px]`}
          >
            <select
              value={selectedDomain}
              onChange={(e) => setSelectedDomain(e.target.value)}
              className={`w-full appearance-none ${theme.inputBg} ${theme.heading} py-4 pl-4 pr-10 focus:outline-none cursor-pointer`}
              style={{
                colorScheme: isDark ? 'dark' : 'light'
              }}
            >
              <option value="lcd.sh">lcd.sh</option>
              <option value="mix.lol">mix.lol</option>
            </select>
            <ChevronDown
              size={14}
              className={`absolute right-3 top-1/2 -translate-y-1/2 ${theme.subtext} pointer-events-none`}
            />
          </div>

          {/* URL Input */}
          <input
            type="url"
            placeholder="https://example.com/long-url"
            required
            value={inputUrl}
            onChange={(e) => setInputUrl(e.target.value)}
            className={`flex-1 p-4 ${theme.inputBg} ${theme.heading} border-y sm:border-y border-x sm:border-l border-t-0 sm:border-t ${theme.border} focus:outline-none focus:border-emerald-500 transition-colors placeholder:text-gray-600`}
          />

          {/* Submit Button */}
          <button
            disabled={loading}
            type="submit"
            className={`bg-emerald-500 hover:bg-emerald-400 ${theme.buttonText} px-6 py-4 sm:py-0 font-bold flex items-center justify-center gap-2 transition-colors disabled:opacity-50 disabled:cursor-not-allowed`}
          >
            {loading ? <Loader2 size={18} className="animate-spin" /> : <ArrowRight size={18} />}
            <span className="hidden sm:inline">{loading ? 'Processing' : 'Shorten'}</span>
            <span className="sm:hidden">{loading ? 'Processing' : 'Shorten Link'}</span>
          </button>
        </form>
      ) : (
        <div className={`flex gap-0 relative animate-in fade-in slide-in-from-bottom-2 duration-300`}>
          <div
            className={`flex-1 p-4 ${theme.inputBg} ${theme.heading} border ${theme.accentBorder} flex items-center justify-between`}
          >
            <span className="font-medium truncate mr-4">{shortenedResult}</span>
            <span className="text-xs text-emerald-500 flex items-center gap-1 shrink-0">
              <Check size={12} /> Ready
            </span>
          </div>
          <button
            onClick={handleResultCopy}
            className={`bg-emerald-500 hover:bg-emerald-400 ${theme.buttonText} px-6 font-bold flex items-center gap-2 transition-colors`}
          >
            {resultCopied ? <Check size={18} /> : <Copy size={18} />}
            <span className="hidden sm:inline">{resultCopied ? 'Copied' : 'Copy'}</span>
          </button>
        </div>
      )}
    </div>
  )
}

