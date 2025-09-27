import type { HTMLAttributes, ReactNode } from "react";
import { forwardRef } from "react";

interface ContainerProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
  size?: "sm" | "md" | "lg" | "xl" | "full";
  center?: boolean;
}

const Container = forwardRef<HTMLDivElement, ContainerProps>(
  ({ children, size = "lg", center = true, className = "", ...props }, ref) => {
    const baseClasses = "w-full px-4 sm:px-6 lg:px-8";

    const sizeClasses = {
      sm: "max-w-2xl",
      md: "max-w-4xl",
      lg: "max-w-6xl",
      xl: "max-w-7xl",
      full: "max-w-full",
    };

    const centerClasses = center ? "mx-auto" : "";

    const classes = [baseClasses, sizeClasses[size], centerClasses, className]
      .filter(Boolean)
      .join(" ");

    return (
      <div ref={ref} className={classes} {...props}>
        {children}
      </div>
    );
  }
);

Container.displayName = "Container";

export { Container };
export type { ContainerProps };
