"use client";

import Link from "next/link";
import { useState } from "react";
import { Button } from "@/components/ui/Button";
import { Container } from "@/components/ui/Container";
import { navItems } from "@/data/landing";

export function Navbar() {
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  return (
    <header className="sticky top-0 z-50 bg-base-100/80 backdrop-blur-md border-b border-base-300/20">
      <Container>
        <div className="navbar min-h-16 px-0">
          {/* Logo */}
          <div className="navbar-start">
            <Link href="/" className="btn btn-ghost text-xl font-bold">
              <span className="text-primary">Upwork</span>
              <span className="text-accent">Jobs</span>
            </Link>
          </div>

          {/* Desktop Navigation */}
          <div className="navbar-center hidden lg:flex">
            <ul className="menu menu-horizontal px-1 gap-2">
              {navItems.map((item) => (
                <li key={item.name}>
                  <Link
                    href={item.href}
                    className="font-medium hover:text-primary transition-colors"
                  >
                    {item.name}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          {/* CTA Buttons */}
          <div className="navbar-end gap-2">
            <div className="hidden sm:flex gap-2">
              <Button variant="ghost" size="sm">
                <Link href="/login">Sign In</Link>
              </Button>
              <Button variant="primary" size="sm">
                <Link href="/signup">Start Free Trial</Link>
              </Button>
            </div>

            {/* Mobile Menu Button */}
            <div className="lg:hidden">
              <button
                type="button"
                className="btn btn-ghost btn-circle"
                onClick={() => setIsMenuOpen(!isMenuOpen)}
                aria-label="Toggle menu"
              >
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  {isMenuOpen ? (
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M6 18L18 6M6 6l12 12"
                    />
                  ) : (
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M4 6h16M4 12h16M4 18h16"
                    />
                  )}
                </svg>
              </button>
            </div>
          </div>
        </div>

        {/* Mobile Menu */}
        {isMenuOpen && (
          <div className="lg:hidden border-t border-base-300/20 py-4">
            <ul className="menu menu-vertical gap-2">
              {navItems.map((item) => (
                <li key={item.name}>
                  <Link
                    href={item.href}
                    className="font-medium"
                    onClick={() => setIsMenuOpen(false)}
                  >
                    {item.name}
                  </Link>
                </li>
              ))}
              <li className="mt-4">
                <div className="flex flex-col gap-2">
                  <Button variant="ghost" size="sm">
                    <Link href="/login">Sign In</Link>
                  </Button>
                  <Button variant="primary" size="sm">
                    <Link href="/signup">Start Free Trial</Link>
                  </Button>
                </div>
              </li>
            </ul>
          </div>
        )}
      </Container>
    </header>
  );
}
