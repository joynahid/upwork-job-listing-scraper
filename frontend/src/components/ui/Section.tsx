import type { HTMLAttributes, ReactNode } from "react";
import { forwardRef } from "react";
import { Container } from "./Container";

interface SectionProps extends HTMLAttributes<HTMLElement> {
  children: ReactNode;
  containerSize?: "sm" | "md" | "lg" | "xl" | "full";
  background?: "default" | "muted" | "accent" | "gradient";
  padding?: "sm" | "md" | "lg" | "xl";
}

const Section = forwardRef<HTMLElement, SectionProps>(
  (
    {
      children,
      containerSize = "lg",
      background = "default",
      padding = "lg",
      className = "",
      ...props
    },
    ref
  ) => {
    const baseClasses = "relative";

    const backgroundClasses = {
      default: "",
      muted: "bg-base-200/30",
      accent: "bg-gradient-to-br from-primary/5 to-secondary/5",
      gradient: "bg-gradient-to-br from-base-200/50 via-base-100 to-base-200/30",
    };

    const paddingClasses = {
      sm: "py-8 sm:py-12",
      md: "py-12 sm:py-16",
      lg: "py-16 sm:py-20 lg:py-24",
      xl: "py-20 sm:py-24 lg:py-32",
    };

    const classes = [baseClasses, backgroundClasses[background], paddingClasses[padding], className]
      .filter(Boolean)
      .join(" ");

    return (
      <section ref={ref} className={classes} {...props}>
        <Container size={containerSize}>{children}</Container>
      </section>
    );
  }
);

Section.displayName = "Section";

export { Section };
export type { SectionProps };
