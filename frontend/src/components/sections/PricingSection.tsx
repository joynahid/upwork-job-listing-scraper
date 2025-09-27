import Link from "next/link";
import { Button } from "@/components/ui/Button";
import { Card, CardContent, CardFooter, CardHeader } from "@/components/ui/Card";
import { Icon } from "@/components/ui/Icon";
import { Section } from "@/components/ui/Section";
import { pricingPlans } from "@/data/landing";

export function PricingSection() {
  return (
    <Section id="pricing" padding="lg">
      <div className="text-center mb-16">
        <h2 className="text-3xl sm:text-4xl font-bold mb-4">Simple, Transparent Pricing</h2>
        <p className="text-lg text-base-content/70 max-w-2xl mx-auto">
          Choose the plan that fits your needs. All plans include a 14-day free trial.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 max-w-6xl mx-auto">
        {pricingPlans.map((plan) => (
          <Card
            key={plan.name}
            className={`relative ${
              plan.highlighted
                ? "ring-2 ring-primary scale-105 bg-gradient-to-br from-primary/5 to-secondary/5"
                : ""
            } transition-all duration-300 hover:scale-[1.02]`}
          >
            {plan.highlighted && (
              <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
                <span className="badge badge-primary badge-lg">Most Popular</span>
              </div>
            )}

            <CardHeader>
              <div className="text-center space-y-4">
                <h3 className="text-2xl font-bold">{plan.name}</h3>
                <div className="space-y-1">
                  <div className="flex items-baseline justify-center gap-1">
                    <span className="text-4xl font-bold">{plan.price}</span>
                    <span className="text-base-content/70">/{plan.period}</span>
                  </div>
                  <p className="text-sm text-base-content/70">{plan.description}</p>
                </div>
              </div>
            </CardHeader>

            <CardContent>
              <ul className="space-y-3">
                {plan.features.map((feature) => (
                  <li key={feature} className="flex items-start gap-3">
                    <div className="w-5 h-5 rounded-full bg-success/20 flex items-center justify-center flex-shrink-0 mt-0.5">
                      <svg
                        className="w-3 h-3 text-success"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M5 13l4 4L19 7"
                        />
                      </svg>
                    </div>
                    <span className="text-sm text-base-content/80">{feature}</span>
                  </li>
                ))}
              </ul>
            </CardContent>

            <CardFooter>
              <Button
                variant={plan.highlighted ? "primary" : "outline"}
                size="lg"
                className="w-full"
              >
                <Link href={plan.ctaLink}>{plan.ctaText}</Link>
              </Button>
            </CardFooter>
          </Card>
        ))}
      </div>

      {/* FAQ Link */}
      <div className="text-center mt-12">
        <p className="text-base-content/70 mb-4">Have questions about our pricing?</p>
        <Button variant="ghost">
          <Link href="#faq">View FAQ →</Link>
        </Button>
      </div>

      {/* Money Back Guarantee */}
      <div className="mt-16 text-center p-6 bg-success/10 rounded-2xl border border-success/20">
        <div className="flex items-center justify-center gap-2 mb-2">
          <Icon name="currency-dollar" className="text-success" size="lg" />
          <h4 className="text-lg font-semibold">30-Day Money Back Guarantee</h4>
        </div>
        <p className="text-sm text-base-content/70">
          Not satisfied? Get a full refund within 30 days, no questions asked.
        </p>
      </div>
    </Section>
  );
}
