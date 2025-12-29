import { Header } from "@/components/Header";
import { Footer } from "@/components/Footer";

export default function DocsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="p-6 md:p-12 lg:p-16 flex flex-col items-center transition-colors duration-300 min-h-screen">
      <div className="max-w-5xl w-full space-y-12">
        <Header />
        {children}

        <div className="pt-12">
          <Footer />
        </div>
      </div>
    </div>
  );
}

