import { BenefitsSection } from "@/components/sections/BenefitsSection";
import { CTASection } from "@/components/sections/CTASection";
import { FAQSection } from "@/components/sections/FAQSection";
import { HeroSection } from "@/components/sections/HeroSection";
import { HowItWorksSection } from "@/components/sections/HowItWorksSection";
import { PartnersSection } from "@/components/sections/PartnersSection";
import { PricingSection } from "@/components/sections/PricingSection";
import { TestimonialsSection } from "@/components/sections/TestimonialsSection";

export default function Home() {
  return (
    <main>
      <HeroSection />
      <PartnersSection />
      <BenefitsSection />
      <HowItWorksSection />
      <PricingSection />
      <TestimonialsSection />
      <FAQSection />
      <CTASection />
    </main>
  );
}
