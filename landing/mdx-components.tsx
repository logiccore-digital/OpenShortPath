import type { MDXComponents } from "mdx/types";
import { CodeBlock } from "@/components/docs/CodeBlock";

export function useMDXComponents(components: MDXComponents): MDXComponents {
  return {
    // Use our custom CodeBlock component for code blocks
    pre: ({ children, ...props }: any) => {
      const codeProps = children?.props || {};
      return (
        <CodeBlock
          className={codeProps.className}
          language={codeProps.language}
        >
          {codeProps.children || children}
        </CodeBlock>
      );
    },
    code: ({ children, className, ...props }: any) => {
      // If it's inside a pre tag, it will be handled by the pre component
      // Otherwise, render inline code
      if (className) {
        return (
          <code className={className} {...props}>
            {children}
          </code>
        );
      }
      return (
        <code
          className="px-1.5 py-0.5 rounded bg-gray-800 text-gray-100 text-sm font-mono"
          {...props}
        >
          {children}
        </code>
      );
    },
    ...components,
  };
}

