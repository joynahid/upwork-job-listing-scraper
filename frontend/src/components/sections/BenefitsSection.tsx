import { Card, CardContent, CardHeader } from "@/components/ui/Card";
import { Icon } from "@/components/ui/Icon";
import { Section } from "@/components/ui/Section";
import { benefits } from "@/data/landing";

export function BenefitsSection() {
  return (
    <Section id="features" padding="lg">
      <div className="text-center mb-16">
        <h2 className="text-3xl sm:text-4xl font-bold mb-4">Why Choose Our Job Feed?</h2>
        <p className="text-lg text-base-content/70 max-w-2xl mx-auto">
          Built specifically for automation professionals who need reliable, real-time access to
          high-quality Upwork opportunities.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
        {benefits.map((benefit) => (
          <Card
            key={benefit.title}
            hover
            className={`${
              benefit.highlight
                ? "ring-2 ring-primary/20 bg-gradient-to-br from-primary/5 to-secondary/5"
                : ""
            } transition-all duration-300 hover:scale-105`}
          >
            <CardHeader>
              <div className="flex items-start gap-4">
                <div className="text-primary flex-shrink-0 mt-1">
                  <Icon name={benefit.icon} size="lg" />
                </div>
                <div>
                  <h3 className="text-xl font-semibold mb-2 flex items-center gap-2">
                    {benefit.title}
                    {benefit.highlight && (
                      <span className="badge badge-primary badge-sm">Popular</span>
                    )}
                  </h3>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <p className="text-base-content/70 leading-relaxed">{benefit.description}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Additional Stats */}
      <div className="mt-16 grid grid-cols-1 sm:grid-cols-3 gap-8 text-center">
        <div className="space-y-3">
          <Icon name="clock" className="text-primary mx-auto" size="xl" />
          <div className="text-3xl font-bold text-primary">30s</div>
          <div className="text-sm text-base-content/70">Average notification time</div>
        </div>
        <div className="space-y-3">
          <Icon name="shield" className="text-primary mx-auto" size="xl" />
          <div className="text-3xl font-bold text-primary">99.9%</div>
          <div className="text-sm text-base-content/70">Service uptime</div>
        </div>
        <div className="space-y-3">
          <Icon name="trending-up" className="text-primary mx-auto" size="xl" />
          <div className="text-3xl font-bold text-primary">10k+</div>
          <div className="text-sm text-base-content/70">Jobs processed daily</div>
        </div>
      </div>
    </Section>
  );
}
