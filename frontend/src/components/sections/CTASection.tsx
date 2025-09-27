import Link from "next/link";
import { Button } from "@/components/ui/Button";
import { Icon } from "@/components/ui/Icon";
import { Section } from "@/components/ui/Section";
import { ctaData } from "@/data/landing";

export function CTASection() {
  return (
    <Section padding="lg" background="gradient">
      <div className="text-center space-y-8">
        <div className="space-y-4">
          <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold bg-gradient-to-r from-primary to-accent bg-clip-text text-transparent">
            {ctaData.headline}
          </h2>
          <p className="text-lg text-base-content/70 max-w-2xl mx-auto">{ctaData.description}</p>
        </div>

        <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
          <Button size="lg" variant={ctaData.primaryCTA.variant}>
            <Link href={ctaData.primaryCTA.href}>{ctaData.primaryCTA.text}</Link>
          </Button>
          <Button size="lg" variant={ctaData.secondaryCTA.variant}>
            <Link href={ctaData.secondaryCTA.href}>{ctaData.secondaryCTA.text}</Link>
          </Button>
        </div>

        {/* Additional incentives */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-6 mt-12 pt-8 border-t border-base-300/20">
          <div className="flex items-center justify-center gap-3">
            <Icon name="check" className="text-success" />
            <span className="text-sm font-medium">14-day free trial</span>
          </div>
          <div className="flex items-center justify-center gap-3">
            <Icon name="x" className="text-error" />
            <span className="text-sm font-medium">No credit card required</span>
          </div>
          <div className="flex items-center justify-center gap-3">
            <Icon name="clock" className="text-primary" />
            <span className="text-sm font-medium">Setup in 5 minutes</span>
          </div>
        </div>
      </div>
    </Section>
  );
}
