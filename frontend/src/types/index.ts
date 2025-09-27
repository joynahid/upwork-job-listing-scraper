// Job API Types (from Go backend)
export interface JobDTO {
  id: string;
  title?: string;
  description?: string;
  job_type?: number;
  status?: number;
  contractor_tier?: number;
  posted_on?: string;
  category?: CategoryInfo;
  budget?: BudgetInfo;
  buyer?: BuyerInfo;
  tags?: string[];
  url?: string;
  last_visited_at?: string;
  duration_label?: string;
  engagement?: string;
  skills?: string[];
  hourly_budget?: HourlyBudget;
  client_activity?: ClientActivity;
  location?: JobLocation;
  is_private?: boolean;
  privacy_reason?: string;
}

export interface JobSummaryDTO {
  id: string;
  title?: string;
  description?: string;
  job_type?: number;
  duration_label?: string;
  engagement?: string;
  skills?: string[];
  hourly_budget?: HourlyBudget;
  fixed_budget?: BudgetInfo;
  weekly_budget?: BudgetInfo;
  client?: JobSummaryClient;
  ciphertext?: string;
  url?: string;
  published_on?: string;
  renewed_on?: string;
  last_visited_at?: string;
}

export interface BudgetInfo {
  fixed_amount?: number;
  currency?: string;
}

export interface HourlyBudget {
  min?: number;
  max?: number;
  currency?: string;
}

export interface CategoryInfo {
  name?: string;
  slug?: string;
  group?: string;
  group_slug?: string;
}

export interface BuyerInfo {
  payment_verified?: boolean;
  country?: string;
  city?: string;
  timezone?: string;
  total_spent?: number;
  total_assignments?: number;
  total_jobs_with_hires?: number;
}

export interface ClientActivity {
  total_applicants?: number;
  total_hired?: number;
  total_invited_to_interview?: number;
  unanswered_invites?: number;
  invitations_sent?: number;
  last_buyer_activity?: string;
}

export interface JobLocation {
  country?: string;
  city?: string;
  timezone?: string;
}

export interface JobSummaryClient {
  payment_verified?: boolean;
  country?: string;
}

export interface JobsResponse {
  success: boolean;
  data: JobDTO[];
  count: number;
  last_updated: string;
  message?: string;
}

export interface JobListResponse {
  success: boolean;
  data: JobSummaryDTO[];
  count: number;
  last_updated: string;
  message?: string;
}

// Landing Page Types
export interface Feature {
  icon: string;
  title: string;
  description: string;
}

export interface Benefit {
  icon: string;
  title: string;
  description: string;
  highlight?: boolean;
}

export interface Step {
  number: number;
  title: string;
  description: string;
}

export interface PricingPlan {
  name: string;
  price: string;
  period: string;
  description: string;
  features: string[];
  highlighted?: boolean;
  ctaText: string;
  ctaLink: string;
}

export interface Testimonial {
  name: string;
  role: string;
  company: string;
  avatar: string;
  content: string;
  rating: number;
}

export interface FAQ {
  question: string;
  answer: string;
}

export interface Partner {
  name: string;
  logo: string;
  url?: string;
}

export interface SocialLink {
  name: string;
  url: string;
  icon: string;
}

export interface NavItem {
  name: string;
  href: string;
}

export interface CTAButton {
  text: string;
  href: string;
  variant: "primary" | "secondary" | "outline";
}
