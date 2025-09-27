import type { SVGProps } from "react";

interface IconProps extends SVGProps<SVGSVGElement> {
  name: string;
  size?: "sm" | "md" | "lg" | "xl";
}

export function Icon({ name, size = "md", className = "", ...props }: IconProps) {
  const sizeClasses = {
    sm: "w-4 h-4",
    md: "w-5 h-5",
    lg: "w-6 h-6",
    xl: "w-8 h-8",
  };

  const iconClass = `${sizeClasses[size]} ${className}`;

  switch (name) {
    case "lightning":
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M13 10V3L4 14h7v7l9-11h-7z"
          />
        </svg>
      );

    case "target":
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <circle cx="12" cy="12" r="10" strokeWidth={2} />
          <circle cx="12" cy="12" r="6" strokeWidth={2} />
          <circle cx="12" cy="12" r="2" strokeWidth={2} />
        </svg>
      );

    case "link":
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.102m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
          />
        </svg>
      );

    case "chart":
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
          />
        </svg>
      );

    case "shield":
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
          />
        </svg>
      );

    case "cog":
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
          />
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
          />
        </svg>
      );

    case "refresh":
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
          />
        </svg>
      );

    case "filter":
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
          />
        </svg>
      );

    case "code":
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"
          />
        </svg>
      );

    case "device-mobile":
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 18h.01M8 21h8a1 1 0 001-1V4a1 1 0 00-1-1H8a1 1 0 00-1 1v16a1 1 0 001 1z"
          />
        </svg>
      );

    case "check":
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
        </svg>
      );

    case "x":
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M6 18L18 6M6 6l12 12"
          />
        </svg>
      );

    case "clock":
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <circle cx="12" cy="12" r="10" strokeWidth={2} />
          <polyline points="12,6 12,12 16,14" strokeWidth={2} />
        </svg>
      );

    case "users":
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a4 4 0 11-8 0 4 4 0 018 0z"
          />
        </svg>
      );

    case "trending-up":
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <polyline points="23 6 13.5 15.5 8.5 10.5 1 18" strokeWidth={2} />
          <polyline points="17 6 23 6 23 12" strokeWidth={2} />
        </svg>
      );

    case "support":
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192L5.636 18.364M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z"
          />
        </svg>
      );

    case "currency-dollar":
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1"
          />
        </svg>
      );

    case "rocket":
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
          />
        </svg>
      );

    case "lock":
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <rect x="3" y="11" width="18" height="11" rx="2" ry="2" strokeWidth={2} />
          <circle cx="12" cy="16" r="1" strokeWidth={2} />
          <path d="M7 11V7a5 5 0 0110 0v4" strokeWidth={2} />
        </svg>
      );

    default:
      return (
        <svg className={iconClass} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
          <circle cx="12" cy="12" r="10" strokeWidth={2} />
          <path d="M9.09 9a3 3 0 015.83 1c0 2-3 3-3 3" strokeWidth={2} />
          <path d="M12 17h.01" strokeWidth={2} />
        </svg>
      );
  }
}

export type { IconProps };
