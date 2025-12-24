import { notFound } from "next/navigation";
import { readFileSync, readdirSync } from "fs";
import { join } from "path";
import React from "react";
import dynamic from "next/dynamic";
import { MDXRemote } from "next-mdx-remote/rsc";
import Link from "next/link";
import { ChevronLeft } from "lucide-react";
import { TableOfContents } from "@/components/docs/TableOfContents";

const CodeBlockWrapper = dynamic(
  () => import("@/components/docs/CodeBlockWrapper").then((mod) => ({ default: mod.CodeBlockWrapper })),
  { ssr: false }
);

const docsDirectory = join(process.cwd(), "content/docs");

// Helper to generate slug from text
function slugify(text: string) {
  return text
    .toLowerCase()
    .replace(/[^\w\s-]/g, "")
    .replace(/\s+/g, "-");
}

// Extract headings for TOC
function extractHeadings(content: string) {
  const lines = content.split("\n");
  const headings: { id: string; text: string; level: number }[] = [];
  
  // Skip the first H1 if it's the title
  let foundFirstH1 = false;

  for (const line of lines) {
    const match = line.match(/^(#{1,3})\s+(.*)$/);
    if (match) {
      const level = match[1].length;
      const text = match[2].trim();
      
      if (level === 1 && !foundFirstH1) {
        foundFirstH1 = true;
        continue;
      }
      
      headings.push({
        id: slugify(text),
        text,
        level,
      });
    }
  }
  return headings;
}

// Get all available doc slugs
function getDocSlugs(): string[] {
  try {
    const files = readdirSync(docsDirectory);
    return files
      .filter((file) => file.endsWith(".mdx"))
      .map((file) => file.replace(/\.mdx$/, ""));
  } catch {
    return [];
  }
}

// Read a doc file
function getDocContent(slug: string): string | null {
  try {
    const filePath = join(docsDirectory, `${slug}.mdx`);
    return readFileSync(filePath, "utf-8");
  } catch {
    return null;
  }
}

export async function generateStaticParams() {
  const slugs = getDocSlugs();
  return slugs.map((slug) => ({ slug }));
}

export default async function DocPage({
  params,
}: {
  params: { slug: string };
}) {
  const content = getDocContent(params.slug);

  if (!content) {
    notFound();
  }

  const headings = extractHeadings(content);

  const mdxContent = await MDXRemote({
    source: content,
    options: {
      mdxOptions: {},
    },
    components: {
      h1: ({ children, ...props }: any) => (
        <h1 id={slugify(String(children))} {...props}>
          {children}
        </h1>
      ),
      h2: ({ children, ...props }: any) => (
        <h2 id={slugify(String(children))} {...props}>
          {children}
        </h2>
      ),
      h3: ({ children, ...props }: any) => (
        <h3 id={slugify(String(children))} {...props}>
          {children}
        </h3>
      ),
      h4: ({ children, ...props }: any) => (
        <h4 id={slugify(String(children))} {...props}>
          {children}
        </h4>
      ),
      p: ({ children, ...props }: any) => (
        <p {...props}>
          {children}
        </p>
      ),
      ul: ({ children, ...props }: any) => (
        <ul {...props}>
          {children}
        </ul>
      ),
      ol: ({ children, ...props }: any) => (
        <ol {...props}>
          {children}
        </ol>
      ),
      li: ({ children, ...props }: any) => (
        <li {...props}>
          {children}
        </li>
      ),
      strong: ({ children, ...props }: any) => (
        <strong {...props}>
          {children}
        </strong>
      ),
      em: ({ children, ...props }: any) => (
        <em {...props}>
          {children}
        </em>
      ),
      a: ({ children, href, ...props }: any) => (
        <a
          href={href}
          {...props}
        >
          {children}
        </a>
      ),
      blockquote: ({ children, ...props }: any) => (
        <blockquote {...props}>
          {children}
        </blockquote>
      ),
      hr: ({ ...props }: any) => (
        <hr {...props} />
      ),
      pre: (props: any) => {
        const { children, ...rest } = props;
        // MDX wraps code blocks in <pre><code> structure
        const childrenArray = React.Children.toArray(children);
        
        // Find the code element - check all children
        let codeElement: any = null;
        for (const child of childrenArray) {
          if (React.isValidElement(child)) {
            const childType = (child as any).type;
            const childProps = (child as any).props || {};
            
            // Check if it's a code element by type or by className
            if (childType === 'code' || 
                childType?.displayName === 'code' ||
                childProps?.className?.startsWith('language-') ||
                (typeof childType === 'string' && childType === 'code')) {
              codeElement = child;
              break;
            }
          }
        }
        
        // If no code element found, check if children is directly a code element
        if (!codeElement && childrenArray.length === 1 && React.isValidElement(childrenArray[0])) {
          const firstChild = childrenArray[0] as any;
          if (firstChild.type === 'code' || firstChild.props?.className?.startsWith('language-')) {
            codeElement = firstChild;
          }
        }
        
        // Extract code content from code element or directly from pre
        let codeContent = '';
        let className = 'language-text';
        let language = 'text';
        
        if (codeElement) {
          const codeProps = (codeElement as any).props || {};
          className = codeProps.className || 'language-text';
          language = className.replace('language-', '') || 'text';
          
          const codeChildren = codeProps.children;
          
          if (typeof codeChildren === 'string') {
            codeContent = codeChildren;
          } else if (codeChildren !== undefined && codeChildren !== null) {
            // Recursively extract text
            const extractText = (node: any): string => {
              if (node === null || node === undefined) return '';
              if (typeof node === 'string') return node;
              if (typeof node === 'number') return String(node);
              if (Array.isArray(node)) {
                return node.map(extractText).join('');
              }
              if (React.isValidElement(node)) {
                const nodeProps = (node as any).props;
                if (nodeProps?.children !== undefined) {
                  return extractText(nodeProps.children);
                }
              }
              return '';
            };
            codeContent = extractText(codeChildren);
          }
        } else {
          // No code element found - extract directly from pre children
          const extractText = (node: any): string => {
            if (node === null || node === undefined) return '';
            if (typeof node === 'string') return node;
            if (typeof node === 'number') return String(node);
            if (Array.isArray(node)) {
              return node.map(extractText).join('');
            }
            if (React.isValidElement(node)) {
              const nodeProps = (node as any).props;
              if (nodeProps?.children !== undefined) {
                return extractText(nodeProps.children);
              }
            }
            return '';
          };
          codeContent = extractText(children);
        }
        
        // Always use CodeBlockWrapper for all pre elements
        if (codeContent.trim()) {
          return (
            <div className="my-6 not-prose">
              <CodeBlockWrapper
                className={className}
                language={language}
              >
                {codeContent}
              </CodeBlockWrapper>
            </div>
          );
        }
        
        // Fallback - render as regular pre (shouldn't happen, but just in case)
        return <pre className="mb-4 p-4 bg-gray-900 dark:bg-black rounded-lg overflow-x-auto border border-gray-200 dark:border-gray-800 text-gray-800 dark:text-gray-200" {...rest}>{children}</pre>;
      },
      code: ({ children, className, ...props }: any) => {
        // Inline code (no language class)
        if (!className || !className.startsWith('language-')) {
          return (
            <code
              {...props}
            >
              {children}
            </code>
          );
        }
        // Code block code (handled by pre component, but we need to preserve it)
        return <code className={className} {...props}>{children}</code>;
      },
    },
  });

  return (
    <div className="flex flex-col md:flex-row gap-12 py-8">
      {/* Main Content */}
      <main className="flex-1 min-w-0">
        <div className="mb-8">
          <Link
            href="/docs"
            className="inline-flex items-center gap-2 text-sm text-gray-500 hover:text-emerald-500 transition-colors"
          >
            <ChevronLeft size={16} />
            Back to Documentation
          </Link>
        </div>
        <article className="prose prose-emerald dark:prose-invert max-w-none">
          {mdxContent}
        </article>
      </main>

      {/* Table of Contents / Sidebar */}
      <aside className="hidden lg:block w-64 flex-shrink-0">
        <div className="sticky top-8 space-y-8">
          <TableOfContents headings={headings} />
          
          <div className="pt-8 border-t border-gray-100 dark:border-gray-800 space-y-4">
            <h4 className="text-sm font-semibold text-gray-900 dark:text-gray-100 uppercase tracking-wider">
              Resources
            </h4>
            <ul className="space-y-2">
              <li>
                <a href="https://github.com/your-repo/OpenShortPath" className="text-sm text-gray-500 hover:text-emerald-500 transition-colors">
                  GitHub Repository
                </a>
              </li>
              <li>
                <Link href="/docs/api-reference" className="text-sm text-gray-500 hover:text-emerald-500 transition-colors">
                  API Reference
                </Link>
              </li>
            </ul>
          </div>
        </div>
      </aside>
    </div>
  );
}

