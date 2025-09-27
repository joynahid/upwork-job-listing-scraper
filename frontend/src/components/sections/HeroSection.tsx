"use client";

import Link from "next/link";
import { Button } from "@/components/ui/Button";
import { Section } from "@/components/ui/Section";
import { heroData } from "@/data/landing";

const heroHighlights = [
  "Stable IDs baked in for versioning and deduplication",
  "Buyer and budget context before you write the first line",
  "Skills and tags ready for clustering, prompts, or outreach",
];

const sampleJobResponse = `{
  "success": true,
  "data": [
    {
      "id": "upwork-872341",
      "title": "Launch a weekly AI founder newsletter",
      "posted_on": "2024-10-24T08:12:43Z",
      "category": {
        "name": "Writing & Translation",
        "group": "Sales & Marketing"
      },
      "budget": {
        "fixed_amount": 2500,
        "currency": "USD"
      },
      "buyer": {
        "payment_verified": true,
        "country": "US"
      },
      "skills": ["newsletter", "ai research", "marketing"],
      "tags": ["founder stories", "growth marketing"]
    }
  ],
  "count": 1,
  "last_updated": "2024-10-24T08:14:03Z"
}`;

export function HeroSection() {
  return (
    <Section padding="xl">
      <div className="grid gap-14 lg:grid-cols-[minmax(0,1fr)_minmax(0,440px)] items-start">
        <div className="space-y-8">
          <div className="space-y-6">
            <span className="inline-flex items-center rounded-full bg-primary/10 px-4 py-1 text-sm font-medium text-primary">
              Upwork Job Schema · API
            </span>
            <h1 className="text-4xl font-semibold leading-tight text-base-content sm:text-5xl">
              {heroData.headline}
            </h1>
            <p className="text-lg text-base-content/80 sm:text-xl">{heroData.subheadline}</p>
            <p className="text-base text-base-content/70 sm:text-lg max-w-xl">
              {heroData.description}
            </p>
          </div>

          <ul className="space-y-3 text-base text-base-content/80">
            {heroHighlights.map((item) => (
              <li key={item} className="flex gap-3">
                <span className="mt-1 inline-block h-1.5 w-1.5 flex-none rounded-full bg-primary"></span>
                <span>{item}</span>
              </li>
            ))}
          </ul>

          <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
            <Button size="lg" variant={heroData.primaryCTA.variant}>
              <Link href={heroData.primaryCTA.href}>{heroData.primaryCTA.text}</Link>
            </Button>
            <Button size="lg" variant="outline">
              <Link href={heroData.secondaryCTA.href}>{heroData.secondaryCTA.text}</Link>
            </Button>
          </div>

          <p className="text-sm text-base-content/60">
            Ship your first automation in minutes with clean, dependable data.
          </p>
        </div>

        <div className="relative">
          <div className="rounded-3xl border border-base-300 bg-base-200/60 p-6 shadow-xl">
            <div className="mb-4 flex items-center gap-3 text-sm text-base-content/60">
              <span className="inline-flex h-2.5 w-2.5 rounded-full bg-success"></span>
              <span>Latest response</span>
            </div>
            <pre className="max-h-[420px] overflow-auto rounded-2xl bg-neutral-900 p-6 text-left text-sm leading-relaxed text-neutral-100">
              <code>{sampleJobResponse}</code>
            </pre>
          </div>
        </div>
      </div>
    </Section>
  );
}
