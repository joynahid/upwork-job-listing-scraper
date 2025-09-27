"use client";

import { useState } from "react";
import { Button } from "@/components/ui/Button";
import { Card, CardContent } from "@/components/ui/Card";
import { Icon } from "@/components/ui/Icon";
import { Section } from "@/components/ui/Section";
import { testimonials } from "@/data/landing";

export function TestimonialsSection() {
  const [activeIndex, setActiveIndex] = useState(0);

  const nextTestimonial = () => {
    setActiveIndex((prev) => (prev + 1) % testimonials.length);
  };

  const prevTestimonial = () => {
    setActiveIndex((prev) => (prev - 1 + testimonials.length) % testimonials.length);
  };

  return (
    <Section id="testimonials" padding="lg" background="muted">
      <div className="text-center mb-16">
        <h2 className="text-3xl sm:text-4xl font-bold mb-4">Loved by Professionals Worldwide</h2>
        <p className="text-lg text-base-content/70 max-w-2xl mx-auto">
          See what automation experts and freelancers are saying about our service.
        </p>
      </div>

      {/* Desktop: Grid Layout */}
      <div className="hidden lg:grid grid-cols-1 lg:grid-cols-3 gap-8">
        {testimonials.map((testimonial) => (
          <TestimonialCard key={testimonial.name} testimonial={testimonial} />
        ))}
      </div>

      {/* Mobile: Carousel Layout */}
      <div className="lg:hidden">
        <div className="relative">
          <TestimonialCard testimonial={testimonials[activeIndex]} />

          {/* Navigation */}
          <div className="flex justify-between items-center mt-6">
            <Button variant="ghost" size="sm" onClick={prevTestimonial}>
              ← Previous
            </Button>

            <div className="flex gap-2">
              {testimonials.map((_, index) => (
                <button
                  key={index}
                  type="button"
                  className={`w-2 h-2 rounded-full transition-colors ${
                    index === activeIndex ? "bg-primary" : "bg-base-300"
                  }`}
                  onClick={() => setActiveIndex(index)}
                />
              ))}
            </div>

            <Button variant="ghost" size="sm" onClick={nextTestimonial}>
              Next →
            </Button>
          </div>
        </div>
      </div>

      {/* Trust Indicators */}
      <div className="mt-16 grid grid-cols-1 sm:grid-cols-4 gap-8 text-center">
        <div className="space-y-3">
          <Icon name="check" className="text-primary mx-auto" size="lg" />
          <div className="text-2xl font-bold text-primary">4.9/5</div>
          <div className="text-sm text-base-content/70">Average Rating</div>
        </div>
        <div className="space-y-3">
          <Icon name="users" className="text-primary mx-auto" size="lg" />
          <div className="text-2xl font-bold text-primary">2,500+</div>
          <div className="text-sm text-base-content/70">Happy Customers</div>
        </div>
        <div className="space-y-3">
          <Icon name="trending-up" className="text-primary mx-auto" size="lg" />
          <div className="text-2xl font-bold text-primary">50k+</div>
          <div className="text-sm text-base-content/70">Jobs Delivered</div>
        </div>
        <div className="space-y-3">
          <Icon name="support" className="text-primary mx-auto" size="lg" />
          <div className="text-2xl font-bold text-primary">24/7</div>
          <div className="text-sm text-base-content/70">Support Available</div>
        </div>
      </div>
    </Section>
  );
}

function TestimonialCard({ testimonial }: { testimonial: (typeof testimonials)[0] }) {
  return (
    <Card className="h-full">
      <CardContent>
        <div className="space-y-4">
          {/* Rating Stars */}
          <div className="flex gap-1">
            {Array.from({ length: 5 }).map((_, i) => (
              <span
                key={i}
                className={`text-lg ${i < testimonial.rating ? "text-warning" : "text-base-300"}`}
              >
                ★
              </span>
            ))}
          </div>

          {/* Quote */}
          <blockquote className="text-base-content/80 italic leading-relaxed">
            &quot;{testimonial.content}&quot;
          </blockquote>

          {/* Author */}
          <div className="flex items-center gap-4 pt-4 border-t border-base-300/20">
            <div className="w-12 h-12 rounded-full bg-gradient-to-br from-primary to-secondary flex items-center justify-center text-primary-content font-semibold">
              {testimonial.name
                .split(" ")
                .map((n) => n[0])
                .join("")}
            </div>
            <div>
              <div className="font-semibold">{testimonial.name}</div>
              <div className="text-sm text-base-content/70">
                {testimonial.role} at {testimonial.company}
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
