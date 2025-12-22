# OpenShortPath Landing Page

A Next.js landing page for OpenShortPath, built with React, TypeScript, and Tailwind CSS.

## Getting Started

### Prerequisites

- Node.js 18+
- npm or yarn

### Installation

1. Install dependencies:

```bash
npm install
```

2. Run the development server:

```bash
npm run dev
```

3. Open [http://localhost:3000](http://localhost:3000) in your browser.

### Build for Production

```bash
npm run build
npm start
```

## Project Structure

```
landing/
├── app/                    # Next.js App Router pages and layouts
├── components/            # React components and theme system
└── package.json           # Dependencies and scripts
```

## Technologies

- **Next.js 14**: React framework with App Router
- **TypeScript**: Type safety
- **Tailwind CSS**: Utility-first CSS framework
- **Lucide React**: Icon library
- **React Hooks**: State management

## Customization

The theme system is centralized in `components/theme/ThemeProvider.tsx`. To modify colors or styling, update the theme object in the provider.
