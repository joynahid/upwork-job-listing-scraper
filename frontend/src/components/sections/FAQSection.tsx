import { Section } from "@/components/ui/Section";
import { faqs } from "@/data/landing";
export function FAQSection() {
  return (
    <Section id="faq" padding="lg" background="muted">
      <div className="mx-auto max-w-2xl text-center">
        <h2 className="text-3xl font-semibold text-base-content sm:text-4xl">
          Frequently asked
        </h2>
        <p className="mt-4 text-base text-base-content/70 sm:text-lg">
          Quick answers to the most common questions about connecting to the Upwork job feed.
        </p>
      </div>

      <div className="mx-auto mt-12 max-w-3xl space-y-4">
        {faqs.map((faq, index) => (
          <details
            key={faq.question}
            className="group rounded-2xl border border-base-300 bg-base-100/80 p-6 shadow-sm"
            {...(index === 0 ? { open: true } : {})}
          >
            <summary className="cursor-pointer text-lg font-medium text-base-content">
              {faq.question}
            </summary>
            <p className="mt-4 text-base text-base-content/70 leading-relaxed">{faq.answer}</p>
          </details>
        ))}
      </div>

      <div className="mt-10 text-center text-sm text-base-content/60">
        Still looking for something else? <a href="mailto:support@upworkjobs.com" className="link">Email our team</a>.
      </div>
    </Section>
  );
}
