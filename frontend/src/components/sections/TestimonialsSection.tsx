import { Section } from "@/components/ui/Section";
import { testimonials } from "@/data/landing";
const MAX_TESTIMONIALS = 2;

export function TestimonialsSection() {
  const featuredTestimonials = testimonials.slice(0, MAX_TESTIMONIALS);

  return (
    <Section id="testimonials" padding="lg">
      <div className="mx-auto max-w-2xl text-center">
        <h2 className="text-3xl font-semibold text-base-content sm:text-4xl">
          Teams rely on us to surface the right work
        </h2>
        <p className="mt-4 text-base text-base-content/70 sm:text-lg">
          Hear from builders who plugged our feed into their automations and stopped prospecting
          the hard way.
        </p>
      </div>

      <div className="mt-12 grid gap-6 lg:grid-cols-2">
        {featuredTestimonials.map((testimonial) => (
          <figure
            key={testimonial.name}
            className="flex h-full flex-col gap-4 rounded-2xl border border-base-300 bg-base-100/80 p-6 shadow-sm"
          >
            <blockquote className="text-base text-base-content/80 leading-relaxed">
              “{testimonial.content}”
            </blockquote>
            <figcaption className="pt-4 text-sm text-base-content/70">
              <span className="font-semibold text-base-content">{testimonial.name}</span> · {" "}
              {testimonial.role}, {testimonial.company}
            </figcaption>
          </figure>
        ))}
      </div>
    </Section>
  );
}
