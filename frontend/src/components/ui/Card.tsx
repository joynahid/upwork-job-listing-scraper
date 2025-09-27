import type { HTMLAttributes, ReactNode } from "react";
import { forwardRef } from "react";

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
  variant?: "default" | "bordered" | "compact" | "side";
  hover?: boolean;
}

const Card = forwardRef<HTMLDivElement, CardProps>(
  ({ children, variant = "default", hover = false, className = "", ...props }, ref) => {
    const baseClasses = "card bg-base-200/50 backdrop-blur-sm";

    const variantClasses = {
      default: "card-body",
      bordered: "card-bordered card-body",
      compact: "card-compact card-body",
      side: "card-side card-body",
    };

    const hoverClasses = hover
      ? "hover:shadow-lg hover:scale-[1.02] transition-all duration-300"
      : "";

    const cardClasses = [baseClasses, hoverClasses, className].filter(Boolean).join(" ");
    const bodyClasses = variantClasses[variant];

    return (
      <div ref={ref} className={cardClasses} {...props}>
        <div className={bodyClasses}>{children}</div>
      </div>
    );
  }
);

Card.displayName = "Card";

// Card sub-components
interface CardHeaderProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
}

const CardHeader = forwardRef<HTMLDivElement, CardHeaderProps>(
  ({ children, className = "", ...props }, ref) => {
    return (
      <div ref={ref} className={`card-title ${className}`} {...props}>
        {children}
      </div>
    );
  }
);

CardHeader.displayName = "CardHeader";

interface CardContentProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
}

const CardContent = forwardRef<HTMLDivElement, CardContentProps>(
  ({ children, className = "", ...props }, ref) => {
    return (
      <div ref={ref} className={`flex-1 ${className}`} {...props}>
        {children}
      </div>
    );
  }
);

CardContent.displayName = "CardContent";

interface CardFooterProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
}

const CardFooter = forwardRef<HTMLDivElement, CardFooterProps>(
  ({ children, className = "", ...props }, ref) => {
    return (
      <div ref={ref} className={`card-actions justify-end ${className}`} {...props}>
        {children}
      </div>
    );
  }
);

CardFooter.displayName = "CardFooter";

export { Card, CardHeader, CardContent, CardFooter };
export type { CardProps, CardHeaderProps, CardContentProps, CardFooterProps };
