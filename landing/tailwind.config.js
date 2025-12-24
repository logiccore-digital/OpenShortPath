/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class',
  content: [
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['var(--font-open-sans)', 'system-ui', 'sans-serif'],
      },
      colors: {
        'terminal-bg': '#0d0d0d',
        'card-bg': '#111',
        'input-bg': '#0a0a0a',
      },
      typography: (theme) => ({
        DEFAULT: {
          css: {
            color: theme('colors.gray.600'),
            maxWidth: 'none',
            hr: {
              borderColor: theme('colors.gray.200'),
              marginTop: '3em',
              marginBottom: '3em',
            },
            'h1, h2, h3, h4': {
              color: theme('colors.gray.900'),
              fontWeight: '700',
              letterSpacing: '-0.025em',
            },
            h1: {
              fontSize: '2.25rem',
              marginTop: '0',
              marginBottom: '2rem',
            },
            h2: {
              fontSize: '1.5rem',
              marginTop: '2.5rem',
              marginBottom: '1rem',
              paddingBottom: '0.5rem',
              borderBottomWidth: '1px',
              borderBottomColor: theme('colors.gray.100'),
            },
            h3: {
              fontSize: '1.25rem',
              marginTop: '2rem',
              marginBottom: '0.75rem',
            },
            a: {
              color: theme('colors.emerald.600'),
              textDecoration: 'none',
              fontWeight: '500',
              '&:hover': {
                color: theme('colors.emerald.500'),
                textDecoration: 'underline',
              },
            },
            strong: {
              color: theme('colors.gray.900'),
              fontWeight: '600',
            },
            code: {
              color: theme('colors.gray.900'),
              backgroundColor: theme('colors.gray.100'),
              paddingLeft: '0.25rem',
              paddingRight: '0.25rem',
              paddingTop: '0.125rem',
              paddingBottom: '0.125rem',
              borderRadius: '0.25rem',
              fontWeight: '500',
            },
            'code::before': {
              content: '""',
            },
            'code::after': {
              content: '""',
            },
            pre: {
              backgroundColor: theme('colors.gray.900'),
              color: theme('colors.gray.200'),
              borderRadius: '0.5rem',
            },
            blockquote: {
              color: theme('colors.gray.500'),
              borderLeftColor: theme('colors.gray.200'),
              fontStyle: 'italic',
            },
          },
        },
        invert: {
          css: {
            color: theme('colors.gray.300'),
            'h1, h2, h3, h4': {
              color: theme('colors.gray.100'),
            },
            h2: {
              borderBottomColor: theme('colors.gray.800'),
            },
            a: {
              color: theme('colors.emerald.500'),
              '&:hover': {
                color: theme('colors.emerald.400'),
              },
            },
            strong: {
              color: theme('colors.gray.100'),
            },
            code: {
              color: theme('colors.gray.200'),
              backgroundColor: theme('colors.gray.900'),
              borderColor: theme('colors.gray.800'),
              borderWidth: '1px',
            },
            blockquote: {
              color: theme('colors.gray.400'),
              borderLeftColor: theme('colors.gray.800'),
            },
            hr: {
              borderColor: theme('colors.gray.800'),
            },
          },
        },
      }),
    },
  },
  plugins: [
    require('@tailwindcss/typography'),
  ],
}

