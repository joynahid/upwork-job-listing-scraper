import Link from "next/link";
import { Section } from "@/components/ui/Section";
import { partners } from "@/data/landing";

export function PartnersSection() {
  return (
    <Section padding="md" background="muted">
      <div className="text-center">
        <p className="text-sm font-medium text-base-content/70 mb-8">
          Integrates seamlessly with your favorite automation tools
        </p>

        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-8 items-center">
          {partners.map((partner) => (
            <div
              key={partner.name}
              className="flex items-center justify-center p-4 rounded-lg bg-base-100/50 hover:bg-base-100 transition-colors group"
            >
              {partner.url ? (
                <Link
                  href={partner.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center justify-center w-full h-12"
                >
                  <PartnerLogo name={partner.name} />
                </Link>
              ) : (
                <div className="flex items-center justify-center w-full h-12">
                  <PartnerLogo name={partner.name} />
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </Section>
  );
}

// Partner Logo Component (using text for now, can be replaced with actual logos)
function PartnerLogo({ name }: { name: string }) {
  const logoClass =
    "text-base-content/60 group-hover:text-base-content transition-colors font-semibold";

  return <span className={logoClass}>{name}</span>;
}
