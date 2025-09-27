import Link from "next/link";
import { Section } from "@/components/ui/Section";
import { partners } from "@/data/landing";

export function PartnersSection() {
  return (
    <Section padding="md">
      <div className="mx-auto max-w-5xl">
        <p className="text-center text-sm font-medium uppercase tracking-wide text-base-content/50">
          Works with every automation stack
        </p>

        <div className="mt-8 flex flex-wrap items-center justify-center gap-6 sm:gap-10">
          {partners.map((partner) => (
            <Link
              key={partner.name}
              href={partner.url ?? "#"}
              className="text-sm font-semibold text-base-content/50 transition-colors hover:text-base-content"
              target={partner.url ? "_blank" : undefined}
              rel={partner.url ? "noopener noreferrer" : undefined}
            >
              {partner.name}
            </Link>
          ))}
        </div>
      </div>
    </Section>
  );
}
