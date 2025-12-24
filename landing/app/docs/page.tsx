import Link from "next/link";
import { FileText, ArrowRight } from "lucide-react";

// List of available docs (flat structure)
const docs = [
  {
    slug: "api-reference",
    title: "API Reference",
    description: "Complete API documentation for all endpoints, including authentication and rate limits.",
  },
  {
    slug: "plans-and-limits",
    title: "Plans and Limits",
    description: "Comprehensive guide to all available plans, usage limits, anonymous access, and verified access options.",
  },
];

export default function DocsPage() {
  return (
    <div className="space-y-12 py-8">
      <div className="space-y-4">
        <h1 className="text-4xl md:text-5xl font-bold text-gray-900 dark:text-gray-100 tracking-tight">
          Documentation
        </h1>
        <p className="text-xl text-gray-500 max-w-2xl">
          Everything you need to integrate OpenShortPath into your applications and master its features.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {docs.map((doc) => (
          <Link
            key={doc.slug}
            href={`/docs/${doc.slug}`}
            className="group block p-8 border border-gray-200 dark:border-gray-800 hover:border-emerald-500 dark:hover:border-emerald-500 rounded-xl transition-all duration-300 bg-white dark:bg-[#111] hover:shadow-lg dark:hover:shadow-emerald-500/10"
          >
            <div className="flex flex-col h-full space-y-4">
              <div className="p-3 bg-emerald-50 dark:bg-emerald-500/10 rounded-lg w-fit transition-colors group-hover:bg-emerald-100 dark:group-hover:bg-emerald-500/20">
                <FileText size={24} className="text-emerald-600 dark:text-emerald-500" />
              </div>
              <div>
                <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-2 flex items-center gap-2">
                  {doc.title}
                  <ArrowRight size={18} className="opacity-0 -translate-x-2 group-hover:opacity-100 group-hover:translate-x-0 transition-all duration-300 text-emerald-500" />
                </h2>
                <p className="text-gray-500 dark:text-gray-400 leading-relaxed">
                  {doc.description}
                </p>
              </div>
            </div>
          </Link>
        ))}
      </div>

      <div className="pt-12 border-t border-gray-100 dark:border-gray-800">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-6">
          Quick Resources
        </h3>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8">
          <div className="space-y-3">
            <h4 className="text-sm font-bold uppercase tracking-wider text-emerald-600 dark:text-emerald-500">Getting Started</h4>
            <ul className="space-y-2 text-sm text-gray-500 dark:text-gray-400">
              <li className="hover:text-gray-900 dark:hover:text-gray-100 transition-colors pointer-events-none opacity-50">Installation (Coming Soon)</li>
              <li className="hover:text-gray-900 dark:hover:text-gray-100 transition-colors pointer-events-none opacity-50">Configuration (Coming Soon)</li>
            </ul>
          </div>
          <div className="space-y-3">
            <h4 className="text-sm font-bold uppercase tracking-wider text-emerald-600 dark:text-emerald-500">Guides</h4>
            <ul className="space-y-2 text-sm text-gray-500 dark:text-gray-400">
              <li className="hover:text-gray-900 dark:hover:text-gray-100 transition-colors pointer-events-none opacity-50">Custom Domains (Coming Soon)</li>
              <li className="hover:text-gray-900 dark:hover:text-gray-100 transition-colors pointer-events-none opacity-50">Analytics (Coming Soon)</li>
            </ul>
          </div>
          <div className="space-y-3">
            <h4 className="text-sm font-bold uppercase tracking-wider text-emerald-600 dark:text-emerald-500">Community</h4>
            <ul className="space-y-2 text-sm text-gray-500 dark:text-gray-400">
              <li><a href="https://github.com/your-repo/OpenShortPath" className="hover:text-gray-900 dark:hover:text-gray-100 transition-colors">GitHub</a></li>
              <li><a href="#" className="hover:text-gray-900 dark:hover:text-gray-100 transition-colors pointer-events-none opacity-50">Discord (Coming Soon)</a></li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}
