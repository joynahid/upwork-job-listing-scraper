import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { Footer } from "@/components/layout/Footer";
import { Navbar } from "@/components/layout/Navbar";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "UpworkJobs - Real-Time Job Feed for Automation Professionals",
  description:
    "Get instant access to fresh Upwork job postings with smart filtering. Perfect for developers, freelancers, and automation specialists using n8n, Make, or Zapier.",
  keywords: ["upwork", "jobs", "automation", "n8n", "zapier", "make", "freelance", "api"],
  authors: [{ name: "UpworkJobs" }],
  openGraph: {
    title: "UpworkJobs - Real-Time Job Feed for Automation Professionals",
    description:
      "Never miss high-value automation projects again. Get real-time Upwork job notifications with advanced filtering.",
    type: "website",
    locale: "en_US",
  },
  twitter: {
    card: "summary_large_image",
    title: "UpworkJobs - Real-Time Job Feed",
    description: "Real-time Upwork job notifications for automation professionals",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" data-theme="dark">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased min-h-screen flex flex-col`}
      >
        <Navbar />
        <div className="flex-1">{children}</div>
        <Footer />
      </body>
    </html>
  );
}
