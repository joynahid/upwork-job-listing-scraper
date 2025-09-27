import { Card, CardContent, CardHeader } from "@/components/ui/Card";
import { Section } from "@/components/ui/Section";
import { steps } from "@/data/landing";

export function HowItWorksSection() {
  return (
    <Section id="how-it-works" padding="lg" background="muted">
      <div className="text-center mb-16">
        <h2 className="text-3xl sm:text-4xl font-bold mb-4">How It Works</h2>
        <p className="text-lg text-base-content/70 max-w-2xl mx-auto">
          Get started in minutes with our simple 3-step process. No complex setup required.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 relative">
        {steps.map((step, index) => (
          <div key={step.number} className="relative">
            <Card hover className="h-full">
              <CardHeader>
                <div className="flex items-center gap-4 mb-4">
                  <div className="w-12 h-12 rounded-full bg-primary text-primary-content flex items-center justify-center font-bold text-lg">
                    {step.number}
                  </div>
                  <h3 className="text-xl font-semibold">{step.title}</h3>
                </div>
              </CardHeader>
              <CardContent>
                <p className="text-base-content/70 leading-relaxed">{step.description}</p>
              </CardContent>
            </Card>

            {/* Connection Arrow (desktop only) */}
            {index < steps.length - 1 && (
              <div className="hidden lg:block absolute top-1/2 -right-4 transform -translate-y-1/2 z-10">
                <div className="w-8 h-8 text-primary/40">
                  <svg
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    className="w-full h-full"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 5l7 7-7 7"
                    />
                  </svg>
                </div>
              </div>
            )}
          </div>
        ))}
      </div>

      {/* Code Example */}
      <div className="mt-16">
        <div className="bg-base-200/50 backdrop-blur-sm rounded-2xl p-6 border border-base-300/20">
          <h4 className="text-lg font-semibold mb-4 text-center">Example: n8n Integration</h4>
          <div className="bg-base-300/30 rounded-lg p-4 font-mono text-sm overflow-x-auto">
            <div className="text-success">{/* Webhook URL from our service */}</div>
            <div className="text-primary">POST</div>{" "}
            <span className="text-accent">https://api.upworkjobs.com/webhook</span>
            <br />
            <br />
            <div className="text-success">{/* Automatic job filtering in n8n */}</div>
            <div>
              <span className="text-warning">if</span> (job.budget &gt;{" "}
              <span className="text-info">1000</span> && job.skills.includes(
              <span className="text-accent">&quot;automation&quot;</span>)) {"{"}
            </div>
            <div className="ml-4">
              <span className="text-warning">return</span> sendSlackNotification(job);
            </div>
            <div>{"}"}</div>
          </div>
        </div>
      </div>
    </Section>
  );
}
