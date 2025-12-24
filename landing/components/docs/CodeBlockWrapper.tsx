"use client";

import { CodeBlock } from "./CodeBlock";

interface CodeBlockWrapperProps {
  children: string;
  className?: string;
  language?: string;
}

export function CodeBlockWrapper({ children, className, language }: CodeBlockWrapperProps) {
  return (
    <CodeBlock className={className} language={language}>
      {children}
    </CodeBlock>
  );
}

