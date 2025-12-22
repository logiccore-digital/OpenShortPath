"use client";

import { Header } from "@/components/Header";
import { DomainsSection } from "@/components/DomainsSection";
import { CoreCapabilities } from "@/components/CoreCapabilities";
import { SelfHostingSection } from "@/components/SelfHostingSection";
import { PricingSection } from "@/components/PricingSection";
import { Footer } from "@/components/Footer";
import { HeroSection } from "@/components/HeroSection";

export const dynamic = "force-dynamic";

export default function Home() {
  return (
    <div className="p-6 md:p-12 lg:p-24 flex flex-col items-center transition-colors duration-300">
      {/* Main Container - restricts width for readability */}
      <div className="max-w-3xl w-full space-y-16">
        {/* Header / Nav */}
        <Header />

        {/* Hero Section */}
        <HeroSection />

        {/* Domains */}
        <DomainsSection />

        {/* Core Capabilities */}
        <CoreCapabilities />

        {/* Self-Hosting Architecture */}
        <SelfHostingSection />

        {/* Pricing Strategy / Philosophy */}
        <PricingSection />

        {/* Footer */}
        <Footer />
      </div>
    </div>
  );
}
