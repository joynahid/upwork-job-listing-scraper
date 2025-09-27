"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/Button";
import { Icon } from "@/components/ui/Icon";
import { Section } from "@/components/ui/Section";
import { heroData } from "@/data/landing";
import { apiClient, formatBudget, formatTimeAgo } from "@/lib/api";
import type { JobSummaryDTO } from "@/types";

export function HeroSection() {
  const [recentJobs, setRecentJobs] = useState<JobSummaryDTO[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchRecentJobs = async () => {
      try {
        const response = await apiClient.getFreshJobs(5);
        setRecentJobs(response.data);
      } catch (error) {
        console.error("Failed to fetch recent jobs:", error);
        // Use mock data for demo
        setRecentJobs([
          {
            id: "1",
            title: "Build n8n Automation Workflow for E-commerce",
            fixed_budget: { fixed_amount: 1500, currency: "USD" },
            skills: ["n8n", "Automation", "API Integration"],
            published_on: new Date(Date.now() - 1000 * 60 * 30).toISOString(), // 30 min ago
          },
          {
            id: "2",
            title: "Zapier Expert Needed for CRM Integration",
            hourly_budget: { min: 50, max: 75, currency: "USD" },
            skills: ["Zapier", "CRM", "Salesforce"],
            published_on: new Date(Date.now() - 1000 * 60 * 45).toISOString(), // 45 min ago
          },
        ]);
      } finally {
        setIsLoading(false);
      }
    };

    fetchRecentJobs();
  }, []);

  return (
    <Section padding="xl" background="gradient">
      <div className="text-center space-y-8">
        {/* Main Headline */}
        <div className="space-y-4">
          <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold bg-gradient-to-r from-primary to-accent bg-clip-text text-transparent">
            {heroData.headline}
          </h1>
          <p className="text-xl sm:text-2xl text-base-content/80 font-medium">
            {heroData.subheadline}
          </p>
          <p className="text-lg text-base-content/70 max-w-3xl mx-auto">{heroData.description}</p>
        </div>

        {/* CTA Buttons */}
        <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
          <Button size="lg" variant={heroData.primaryCTA.variant}>
            <Link href={heroData.primaryCTA.href}>{heroData.primaryCTA.text}</Link>
          </Button>
          <Button size="lg" variant={heroData.secondaryCTA.variant}>
            <Link href={heroData.secondaryCTA.href}>{heroData.secondaryCTA.text}</Link>
          </Button>
        </div>

        {/* Live Jobs Feed Preview */}
        <div className="mt-16">
          <div className="bg-base-200/50 backdrop-blur-sm rounded-2xl p-6 border border-base-300/20">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-lg font-semibold flex items-center gap-2">
                <span className="w-3 h-3 bg-success rounded-full animate-pulse"></span>
                Live Job Feed
              </h3>
              <span className="text-sm text-base-content/70">
                Updated {formatTimeAgo(new Date().toISOString())}
              </span>
            </div>

            {isLoading ? (
              <div className="space-y-4">
                {[1, 2, 3].map((i) => (
                  <div key={i} className="animate-pulse">
                    <div className="h-4 bg-base-300 rounded w-3/4 mb-2"></div>
                    <div className="h-3 bg-base-300 rounded w-1/2"></div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="space-y-4 text-left">
                {recentJobs.map((job) => (
                  <div
                    key={job.id}
                    className="p-4 bg-base-100/50 rounded-lg border border-base-300/20 hover:border-primary/20 transition-colors"
                  >
                    <h4 className="font-medium text-base-content mb-2 line-clamp-1">{job.title}</h4>
                    <div className="flex flex-wrap items-center gap-4 text-sm text-base-content/70">
                      <span className="font-medium text-success">
                        {job.fixed_budget
                          ? formatBudget(job.fixed_budget)
                          : job.hourly_budget?.min
                            ? `$${job.hourly_budget.min}-${job.hourly_budget.max}/hr`
                            : "Budget TBD"}
                      </span>
                      <span>{formatTimeAgo(job.published_on)}</span>
                      {job.skills && job.skills.length > 0 && (
                        <div className="flex gap-1">
                          {job.skills.slice(0, 3).map((skill) => (
                            <span key={skill} className="badge badge-sm badge-outline">
                              {skill}
                            </span>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}

            <div className="mt-6 text-center">
              <Button variant="outline" size="sm">
                <Link href="#pricing">View All Jobs →</Link>
              </Button>
            </div>
          </div>
        </div>

        {/* Trust Indicators */}
        <div className="mt-12 pt-8 border-t border-base-300/20">
          <p className="text-sm text-base-content/60 mb-4">
            Trusted by automation professionals worldwide
          </p>
          <div className="flex justify-center items-center gap-8 opacity-60">
            <Icon name="lightning" className="text-warning" />
            <span className="text-sm font-medium">99.9% Uptime</span>
            <Icon name="lock" className="text-success" />
            <span className="text-sm font-medium">Enterprise Security</span>
            <Icon name="rocket" className="text-primary" />
            <span className="text-sm font-medium">Real-time Updates</span>
          </div>
        </div>
      </div>
    </Section>
  );
}
