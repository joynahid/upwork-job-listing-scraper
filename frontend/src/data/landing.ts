import type {
  Benefit,
  FAQ,
  Feature,
  NavItem,
  Partner,
  PricingPlan,
  SocialLink,
  Step,
  Testimonial,
} from "@/types";

// Navigation
export const navItems: NavItem[] = [
  { name: "Features", href: "#features" },
  { name: "How it works", href: "#how-it-works" },
  { name: "Pricing", href: "#pricing" },
  { name: "Testimonials", href: "#testimonials" },
  { name: "FAQ", href: "#faq" },
];

// Hero Section
export const heroData = {
  headline: "Fresh Upwork Jobs in Real-Time",
  subheadline: "Never miss high-value automation projects again",
  description:
    "Get instant access to the latest Upwork job postings with our clean, filtered feed. Perfect for developers, freelancers, and automation specialists using n8n, Make, or Zapier.",
  primaryCTA: {
    text: "Start Free Trial",
    href: "#pricing",
    variant: "primary" as const,
  },
  secondaryCTA: {
    text: "View Live Jobs",
    href: "#live-feed",
    variant: "outline" as const,
  },
  videoUrl: "/demo-video.mp4",
  screenshot: "/dashboard-screenshot.png",
};

// Partners Section
export const partners: Partner[] = [
  { name: "n8n", logo: "/partners/n8n.svg", url: "https://n8n.io" },
  { name: "Make", logo: "/partners/make.svg", url: "https://make.com" },
  { name: "Zapier", logo: "/partners/zapier.svg", url: "https://zapier.com" },
  { name: "Integromat", logo: "/partners/integromat.svg" },
  { name: "Pabbly", logo: "/partners/pabbly.svg" },
  { name: "Automate.io", logo: "/partners/automate.svg" },
];

// Benefits Section
export const benefits: Benefit[] = [
  {
    icon: "lightning",
    title: "Real-Time Updates",
    description: "Get job notifications within seconds of posting on Upwork",
    highlight: true,
  },
  {
    icon: "target",
    title: "Smart Filtering",
    description: "Advanced filters for budget, skills, client quality, and more",
  },
  {
    icon: "link",
    title: "API Integration",
    description: "Connect directly to your automation workflows and tools",
  },
  {
    icon: "chart",
    title: "Clean Data",
    description: "Structured, normalized job data ready for your applications",
  },
  {
    icon: "shield",
    title: "Reliable Service",
    description: "99.9% uptime with enterprise-grade infrastructure",
  },
  {
    icon: "cog",
    title: "Automation Ready",
    description: "Built for n8n, Make, Zapier, and custom integrations",
  },
];

// How It Works Section
export const steps: Step[] = [
  {
    number: 1,
    title: "Connect Your Tools",
    description: "Integrate with n8n, Make, Zapier, or use our REST API directly",
  },
  {
    number: 2,
    title: "Set Your Filters",
    description: "Configure job criteria: budget range, skills, client quality, location",
  },
  {
    number: 3,
    title: "Automate Everything",
    description: "Receive instant notifications and automate your proposal workflow",
  },
];

// Pricing Section
export const pricingPlans: PricingPlan[] = [
  {
    name: "Starter",
    price: "$29",
    period: "month",
    description: "Perfect for individual freelancers getting started",
    features: [
      "Up to 100 job alerts/month",
      "Basic filtering",
      "Email notifications",
      "API access",
      "Community support",
    ],
    ctaText: "Start Free Trial",
    ctaLink: "/signup?plan=starter",
  },
  {
    name: "Pro",
    price: "$79",
    period: "month",
    description: "Best for active freelancers and small agencies",
    features: [
      "Unlimited job alerts",
      "Advanced filtering",
      "Real-time webhooks",
      "Priority API access",
      "Slack/Discord integration",
      "Email support",
      "Custom integrations",
    ],
    highlighted: true,
    ctaText: "Start Free Trial",
    ctaLink: "/signup?plan=pro",
  },
  {
    name: "Enterprise",
    price: "$199",
    period: "month",
    description: "For agencies and teams with advanced automation needs",
    features: [
      "Everything in Pro",
      "White-label API",
      "Custom data fields",
      "Dedicated support",
      "SLA guarantee",
      "Custom integrations",
      "Team management",
      "Analytics dashboard",
    ],
    ctaText: "Contact Sales",
    ctaLink: "/contact",
  },
];

// Testimonials Section
export const testimonials: Testimonial[] = [
  {
    name: "Sarah Chen",
    role: "Full-Stack Developer",
    company: "Freelancer",
    avatar: "/testimonials/sarah.jpg",
    content:
      "This service has completely transformed how I find projects. I'm now landing 3x more high-value automation gigs thanks to the real-time alerts.",
    rating: 5,
  },
  {
    name: "Marcus Rodriguez",
    role: "Automation Specialist",
    company: "AutoFlow Agency",
    avatar: "/testimonials/marcus.jpg",
    content:
      "The n8n integration is seamless. We've automated our entire lead qualification process and increased our proposal success rate by 40%.",
    rating: 5,
  },
  {
    name: "Emily Johnson",
    role: "No-Code Developer",
    company: "Zapier Expert",
    avatar: "/testimonials/emily.jpg",
    content:
      "Finally, a clean data source for Upwork jobs! The API is well-documented and the filtering options are exactly what I needed.",
    rating: 5,
  },
];

// FAQ Section
export const faqs: FAQ[] = [
  {
    question: "How quickly do I receive job notifications?",
    answer:
      "Our system monitors Upwork every 30 seconds and sends notifications within 1-2 minutes of a job being posted. Pro users get priority processing for even faster alerts.",
  },
  {
    question: "Can I integrate this with my existing automation tools?",
    answer:
      "Absolutely! We have native integrations for n8n, Make (Integromat), and Zapier. You can also use our REST API with any tool that supports webhooks or HTTP requests.",
  },
  {
    question: "What kind of filtering options are available?",
    answer:
      "You can filter by budget range, hourly rate, client payment verification, location, required skills, job category, posting date, and much more. Pro users get access to advanced filters like client spending history and hire rate.",
  },
  {
    question: "Is there a free trial?",
    answer:
      "Yes! All plans come with a 14-day free trial. No credit card required to start. You can test all features and see if our service fits your workflow.",
  },
  {
    question: "How reliable is the service?",
    answer:
      "We maintain 99.9% uptime with enterprise-grade infrastructure. Our system includes redundancy, monitoring, and automatic failover to ensure you never miss important job postings.",
  },
  {
    question: "Can I cancel anytime?",
    answer:
      "Yes, you can cancel your subscription at any time. There are no long-term contracts or cancellation fees. Your access continues until the end of your current billing period.",
  },
];

// Social Links
export const socialLinks: SocialLink[] = [
  { name: "Twitter", url: "https://twitter.com/upworkjobs", icon: "twitter" },
  { name: "LinkedIn", url: "https://linkedin.com/company/upworkjobs", icon: "linkedin" },
  { name: "GitHub", url: "https://github.com/upworkjobs", icon: "github" },
  { name: "Discord", url: "https://discord.gg/upworkjobs", icon: "discord" },
];

// Features (for features section)
export const features: Feature[] = [
  {
    icon: "refresh",
    title: "Real-Time Sync",
    description: "Jobs appear in your feed within seconds of being posted on Upwork",
  },
  {
    icon: "filter",
    title: "Advanced Filters",
    description: "Filter by budget, skills, client quality, location, and 20+ other criteria",
  },
  {
    icon: "code",
    title: "API First",
    description: "RESTful API with webhooks, perfect for automation and integrations",
  },
  {
    icon: "device-mobile",
    title: "Multi-Platform",
    description: "Works with n8n, Make, Zapier, Slack, Discord, and custom applications",
  },
];

// CTA Section
export const ctaData = {
  headline: "Ready to Never Miss Another High-Value Project?",
  description:
    "Join thousands of successful freelancers and agencies who use our platform to stay ahead of the competition.",
  primaryCTA: {
    text: "Start Your Free Trial",
    href: "/signup",
    variant: "primary" as const,
  },
  secondaryCTA: {
    text: "Schedule Demo",
    href: "/demo",
    variant: "outline" as const,
  },
};

// Footer
export const footerData = {
  company: {
    name: "UpworkJobs",
    description: "Real-time Upwork job feeds for automation professionals",
    logo: "/logo.svg",
  },
  links: {
    product: [
      { name: "Features", href: "#features" },
      { name: "Pricing", href: "#pricing" },
      { name: "API Docs", href: "/docs" },
      { name: "Integrations", href: "/integrations" },
    ],
    company: [
      { name: "About", href: "/about" },
      { name: "Blog", href: "/blog" },
      { name: "Careers", href: "/careers" },
      { name: "Contact", href: "/contact" },
    ],
    support: [
      { name: "Help Center", href: "/help" },
      { name: "Community", href: "/community" },
      { name: "Status", href: "/status" },
      { name: "Changelog", href: "/changelog" },
    ],
    legal: [
      { name: "Privacy", href: "/privacy" },
      { name: "Terms", href: "/terms" },
      { name: "Security", href: "/security" },
    ],
  },
  newsletter: {
    title: "Stay Updated",
    description: "Get the latest updates on new features and integrations.",
    placeholder: "Enter your email",
    buttonText: "Subscribe",
  },
};
