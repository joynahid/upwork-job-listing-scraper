import Link from "next/link";
import { Button } from "@/components/ui/Button";
import { Section } from "@/components/ui/Section";
import { ctaData } from "@/data/landing";

export function CTASection() {
  return (
    <Section padding="lg">
      <div className="mx-auto max-w-3xl rounded-3xl border border-base-300 bg-base-100/90 p-10 text-center shadow-lg">
        <h2 className="text-3xl font-semibold text-base-content sm:text-4xl">
          {ctaData.headline}
        </h2>
        <p className="mt-4 text-base text-base-content/70 sm:text-lg">{ctaData.description}</p>

        <div className="mt-8 flex flex-col gap-3 sm:flex-row sm:justify-center">
          <Button size="lg" variant={ctaData.primaryCTA.variant}>
            <Link href={ctaData.primaryCTA.href}>{ctaData.primaryCTA.text}</Link>
          </Button>
          <Button size="lg" variant={ctaData.secondaryCTA.variant}>
            <Link href={ctaData.secondaryCTA.href}>{ctaData.secondaryCTA.text}</Link>
          </Button>
        </div>

        <p className="mt-6 text-sm text-base-content/60">
          14-day trial, cancel anytime. No credit card required.
        </p>
      </div>
    </Section>
  );
}
