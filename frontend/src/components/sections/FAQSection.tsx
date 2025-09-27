"use client";

import { useState } from "react";
import { Section } from "@/components/ui/Section";
import { faqs } from "@/data/landing";

export function FAQSection() {
  const [openIndex, setOpenIndex] = useState<number | null>(0);

  const toggleFAQ = (index: number) => {
    setOpenIndex(openIndex === index ? null : index);
  };

  return (
    <Section id="faq" padding="lg">
      <div className="text-center mb-16">
        <h2 className="text-3xl sm:text-4xl font-bold mb-4">Frequently Asked Questions</h2>
        <p className="text-lg text-base-content/70 max-w-2xl mx-auto">
          Everything you need to know about our service. Can&apos;t find what you&apos;re looking
          for? Contact our support team.
        </p>
      </div>

      <div className="max-w-3xl mx-auto space-y-4">
        {faqs.map((faq, index) => (
          <div
            key={faq.question}
            className="collapse collapse-plus bg-base-200/50 backdrop-blur-sm border border-base-300/20"
          >
            <input
              type="radio"
              name="faq-accordion"
              checked={openIndex === index}
              onChange={() => toggleFAQ(index)}
            />
            <div className="collapse-title text-lg font-medium">{faq.question}</div>
            <div className="collapse-content">
              <p className="text-base-content/70 leading-relaxed">{faq.answer}</p>
            </div>
          </div>
        ))}
      </div>

      {/* Contact Support */}
      <div className="text-center mt-12">
        <div className="bg-base-200/50 backdrop-blur-sm rounded-2xl p-8 border border-base-300/20">
          <h3 className="text-xl font-semibold mb-2">Still have questions?</h3>
          <p className="text-base-content/70 mb-6">
            Our support team is here to help you get the most out of our service.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <a href="mailto:support@upworkjobs.com" className="btn btn-primary">
              Email Support
            </a>
            <a href="/contact" className="btn btn-outline">
              Schedule Call
            </a>
          </div>
        </div>
      </div>
    </Section>
  );
}
