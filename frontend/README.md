# UpworkJobs Landing Page

A modern, responsive landing page for the UpworkJobs real-time job feed service, built with Next.js 15, Tailwind CSS v4, and DaisyUI.

## 🚀 Features

- **Modern Tech Stack**: Next.js 15 with App Router, Tailwind CSS v4, DaisyUI
- **Type Safety**: Fully typed with TypeScript and strict Biome linting
- **Responsive Design**: Mobile-first design with dark theme
- **Real-time Data**: Integrates with Go API backend for live job feeds
- **Performance**: Optimized with Turbopack and static generation
- **Accessibility**: WCAG compliant with semantic HTML

## 🛠 Tech Stack

- **Framework**: Next.js 15 (App Router)
- **Styling**: Tailwind CSS v4 + DaisyUI
- **Language**: TypeScript
- **Linting**: Biome (ESLint + Prettier replacement)
- **Build Tool**: Turbopack
- **Deployment**: Vercel-ready

## 📦 Installation

1. **Install dependencies**:
   ```bash
   npm install
   ```

2. **Set up environment variables**:
   ```bash
   cp env.example .env.local
   # Edit .env.local with your API configuration
   ```

3. **Start development server**:
   ```bash
   npm run dev
   ```

4. **Open in browser**: http://localhost:3000

## 🏗 Project Structure

```
src/
├── app/                    # Next.js App Router
│   ├── layout.tsx         # Root layout with Navbar/Footer
│   ├── page.tsx           # Landing page
│   └── globals.css        # Global styles with Tailwind v4
├── components/
│   ├── ui/                # Reusable UI components
│   │   ├── Button.tsx
│   │   ├── Card.tsx
│   │   ├── Container.tsx
│   │   └── Section.tsx
│   ├── layout/            # Layout components
│   │   ├── Navbar.tsx
│   │   └── Footer.tsx
│   └── sections/          # Landing page sections
│       ├── HeroSection.tsx
│       ├── BenefitsSection.tsx
│       ├── PricingSection.tsx
│       └── ...
├── data/
│   └── landing.ts         # Centralized content (single source of truth)
├── lib/
│   └── api.ts             # API client for Go backend
└── types/
    └── index.ts           # TypeScript definitions
```

## 🎨 Design System

### Colors (Dark Theme)
- **Primary**: `oklch(0.7 0.15 260)` - Blue accent
- **Secondary**: `oklch(0.6 0.12 290)` - Purple accent  
- **Accent**: `oklch(0.8 0.14 120)` - Green accent
- **Base**: Dark grays for backgrounds

### Components
- Built with DaisyUI component library
- Custom variants for branding
- Consistent spacing and typography
- Hover animations and transitions

## 🔧 Development

### Available Scripts

```bash
# Development
npm run dev              # Start dev server with Turbopack
npm run build           # Production build
npm run start           # Start production server

# Code Quality
npm run lint            # Run Biome linter
npm run lint:fix        # Fix linting issues
npm run format          # Format code
npm run format:fix      # Format and fix code
npm run check           # Run all checks
npm run check:fix       # Fix all issues
npm run type-check      # TypeScript type checking
```

### Code Quality

- **Biome**: Modern linter and formatter (replaces ESLint + Prettier)
- **TypeScript**: Strict mode enabled
- **Pre-commit hooks**: Automatic formatting and linting
- **Import organization**: Automatic import sorting

## 🌐 API Integration

The landing page integrates with the Go API backend:

- **Health Check**: `/health`
- **Job List**: `/job-list` (with filtering)
- **Jobs**: `/jobs` (detailed job data)

### API Client Usage

```typescript
import { apiClient } from '@/lib/api';

// Get fresh jobs for hero section
const jobs = await apiClient.getFreshJobs(5);

// Get automation-specific jobs
const autoJobs = await apiClient.getAutomationJobs(20);
```

## 📱 Responsive Design

- **Mobile**: 320px+ (stacked layout)
- **Tablet**: 768px+ (2-column grid)
- **Desktop**: 1024px+ (3-column grid)
- **Large**: 1920px+ (max-width container)

## 🚀 Deployment

### Vercel (Recommended)

1. **Connect repository** to Vercel
2. **Set environment variables**:
   - `NEXT_PUBLIC_API_URL`
   - `NEXT_PUBLIC_API_KEY`
3. **Deploy** automatically on push

### Manual Deployment

```bash
npm run build
npm run start
```

## 🎯 Performance

- **First Load JS**: ~124KB (optimized)
- **Static Generation**: Pre-rendered at build time
- **Image Optimization**: Next.js automatic optimization
- **Code Splitting**: Automatic route-based splitting
- **Turbopack**: Fast development builds

## 🔒 Environment Variables

```bash
# Required
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_API_KEY=your-api-key

# Optional
NEXT_PUBLIC_GA_ID=G-XXXXXXXXXX
NEXT_PUBLIC_HOTJAR_ID=1234567
```

## 📝 Content Management

All content is centralized in `src/data/landing.ts`:

- **Hero section**: Headlines, CTAs, descriptions
- **Features**: Benefits and feature lists
- **Pricing**: Plans, features, pricing
- **Testimonials**: Customer reviews
- **FAQ**: Questions and answers
- **Footer**: Links and company info

This ensures easy content updates without touching component code.

## 🤝 Contributing

1. **Follow code style**: Biome configuration
2. **Type everything**: No `any` types
3. **Test components**: Ensure responsive design
4. **Update content**: Use centralized data files
5. **Performance**: Keep bundle size optimized

## 📄 License

This project is part of the UpworkJobs platform. All rights reserved.