import Link from "next/link";
import { Button } from "@/components/ui/Button";
import { Section } from "@/components/ui/Section";
import { pricingPlans } from "@/data/landing";

export function PricingSection() {
  return (
    <Section id="pricing" padding="lg" background="muted">
      <div className="mx-auto max-w-2xl text-center">
        <h2 className="text-3xl font-semibold text-base-content sm:text-4xl">
          Pricing that scales with your pipeline
        </h2>
        <p className="mt-4 text-base text-base-content/70 sm:text-lg">
          Start free, then pick the level of volume and support that suits your team. Cancel
          anytime.
        </p>
      </div>

      <div className="mt-12 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {pricingPlans.map((plan) => {
          const isHighlighted = Boolean(plan.highlighted);

          return (
            <div
              key={plan.name}
              className={`flex h-full flex-col rounded-2xl border p-8 shadow-sm transition-colors ${
                isHighlighted
                  ? "border-primary bg-base-100 shadow-md"
                  : "border-base-300 bg-base-100/80"
              }`}
            >
              <div className="space-y-3 text-left">
                <h3 className="text-2xl font-semibold text-base-content">{plan.name}</h3>
                <div className="flex items-baseline gap-2 text-base-content">
                  <span className="text-4xl font-semibold">{plan.price}</span>
                  <span className="text-base-content/60">/{plan.period}</span>
                </div>
                <p className="text-sm text-base-content/70">{plan.description}</p>
              </div>

              <ul className="mt-8 space-y-3 text-sm text-base-content/80">
                {plan.features.map((feature) => (
                  <li key={feature} className="flex gap-3">
                    <span className="mt-1 inline-block h-1.5 w-1.5 flex-none rounded-full bg-primary"></span>
                    <span>{feature}</span>
                  </li>
                ))}
              </ul>

              <div className="mt-auto pt-8">
                <Button
                  variant={isHighlighted ? "primary" : "outline"}
                  size="lg"
                  className="w-full justify-center"
                >
                  <Link href={plan.ctaLink}>{plan.ctaText}</Link>
                </Button>
              </div>
            </div>
          );
        })}
      </div>

      <div className="mt-10 text-center text-sm text-base-content/60">
        Need something custom? <Link href="/contact" className="link">Talk to us →</Link>
      </div>
    </Section>
  );
}
