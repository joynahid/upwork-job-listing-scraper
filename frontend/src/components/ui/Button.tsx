import type { ButtonHTMLAttributes, ReactNode } from "react";
import { forwardRef } from "react";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "outline" | "ghost";
  size?: "sm" | "md" | "lg";
  children: ReactNode;
  loading?: boolean;
}

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = "primary", size = "md", children, loading, className = "", ...props }, ref) => {
    const baseClasses = "btn font-medium transition-all duration-200 ease-snappy";

    const variantClasses = {
      primary: "btn-primary",
      secondary: "btn-secondary",
      outline: "btn-outline",
      ghost: "btn-ghost",
    };

    const sizeClasses = {
      sm: "btn-sm",
      md: "",
      lg: "btn-lg",
    };

    const classes = [
      baseClasses,
      variantClasses[variant],
      sizeClasses[size],
      loading && "loading",
      className,
    ]
      .filter(Boolean)
      .join(" ");

    return (
      <button ref={ref} className={classes} disabled={loading} {...props}>
        {loading ? (
          <>
            <span className="loading loading-spinner loading-sm"></span>
            Loading...
          </>
        ) : (
          children
        )}
      </button>
    );
  }
);

Button.displayName = "Button";

export { Button };
export type { ButtonProps };
