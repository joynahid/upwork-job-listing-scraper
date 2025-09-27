import { Icon } from "@/components/ui/Icon";
import { Section } from "@/components/ui/Section";
import { benefits } from "@/data/landing";

const FEATURE_COUNT = 3;

export function BenefitsSection() {
  const featuredBenefits = benefits.slice(0, FEATURE_COUNT);

  return (
    <Section id="features" padding="lg" background="muted">
      <div className="mx-auto max-w-2xl text-center">
        <h2 className="text-3xl font-semibold text-base-content sm:text-4xl">
          Ready-to-use data without the cleanup
        </h2>
        <p className="mt-4 text-base text-base-content/70 sm:text-lg">
          Every record is normalized, enriched, and easy to plug into your workflows. No more
          parsing HTML or wrangling inconsistent fields.
        </p>
      </div>

      <div className="mt-12 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {featuredBenefits.map((benefit) => (
          <div
            key={benefit.title}
            className="flex h-full flex-col gap-4 rounded-2xl border border-base-300 bg-base-100/70 p-6 text-left shadow-sm"
          >
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/10 text-primary">
              <Icon name={benefit.icon} />
            </div>
            <div className="space-y-3">
              <h3 className="text-xl font-semibold text-base-content">{benefit.title}</h3>
              <p className="text-base text-base-content/70">{benefit.description}</p>
            </div>
          </div>
        ))}
      </div>
    </Section>
  );
}
