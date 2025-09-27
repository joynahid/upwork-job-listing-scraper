import { Section } from "@/components/ui/Section";
import { steps } from "@/data/landing";

export function HowItWorksSection() {
  return (
    <Section id="how-it-works" padding="lg">
      <div className="mx-auto max-w-2xl text-center">
        <h2 className="text-3xl font-semibold text-base-content sm:text-4xl">
          From new job to shipped workflow in three steps
        </h2>
        <p className="mt-4 text-base text-base-content/70 sm:text-lg">
          Connect once, set the filters that matter, and stream the work straight into your
          automations.
        </p>
      </div>

      <div className="mt-12 grid gap-8 lg:grid-cols-3">
        {steps.map((step) => (
          <div key={step.number} className="flex flex-col gap-4">
            <div className="h-10 w-10 rounded-full bg-primary text-center text-lg font-semibold leading-10 text-primary-content">
              {step.number}
            </div>
            <div className="space-y-3">
              <h3 className="text-xl font-semibold text-base-content">{step.title}</h3>
              <p className="text-base text-base-content/70">{step.description}</p>
            </div>
          </div>
        ))}
      </div>
    </Section>
  );
}
