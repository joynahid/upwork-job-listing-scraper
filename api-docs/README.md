# Upwork Job API Documentation

This directory contains the Docusaurus-powered documentation site for the Upwork Job API.

## 🚀 Quick Start

```bash
# Install dependencies
npm install

# Start development server
npm start

# Build for production
npm run build

# Serve production build locally
npm run serve
```

## 📁 Project Structure

```
api-docs/
├── docs/                          # Documentation content
│   ├── overview.md                # API overview and introduction
│   ├── getting-started.md         # Setup and onboarding guide
│   ├── endpoints/                 # Endpoint-specific documentation
│   │   ├── health.md             # GET /health endpoint
│   │   └── jobs.md               # GET /jobs endpoint
│   └── reference/                 # Reference materials
│       ├── authentication.md     # Authentication guide
│       ├── filter-parameters.md  # Query parameter reference
│       └── docs-site-maintenance.md # Site maintenance guide
├── src/                          # React components and styling
│   ├── components/               # Custom React components
│   ├── css/                      # Global styles
│   └── pages/                    # Custom pages
├── static/                       # Static assets
├── docusaurus.config.ts          # Docusaurus configuration
└── sidebars.ts                   # Sidebar navigation structure
```

## 🎨 Features

- **Modern Design**: Dark-mode friendly with custom styling inspired by modern documentation sites
- **Responsive Layout**: Works seamlessly on desktop, tablet, and mobile devices
- **Fast Search**: Built-in search functionality across all documentation
- **API-Focused**: Tailored specifically for API documentation with endpoint details, examples, and reference materials
- **TypeScript Support**: Full TypeScript configuration for type safety

## 📝 Content Guidelines

### Adding New Documentation

1. Create new `.md` files in the appropriate directory under `docs/`
2. Add frontmatter with `title` and `sidebar_position`
3. Update `sidebars.ts` if adding new sections
4. Use consistent formatting and tone matching existing content

### Markdown Features

- **Code blocks**: Use triple backticks with language specification
- **Admonitions**: Use `:::info`, `:::warning`, `:::danger` for callouts
- **Tables**: Standard markdown table syntax
- **Links**: Use relative paths for internal links

### Style Guide

- Use sentence case for headings
- Include practical examples for all concepts
- Keep explanations concise but comprehensive
- Use consistent terminology matching the Go API codebase

## 🔧 Configuration

### Key Configuration Files

- `docusaurus.config.ts`: Main site configuration, branding, and navigation
- `sidebars.ts`: Documentation sidebar structure
- `src/css/custom.css`: Custom styling and theme overrides

### Customization

The site uses a custom dark theme with:
- Primary color: `#6c5ce7` (light mode), `#74d1f3` (dark mode)
- Custom typography with Inter font family
- Modern card-based layouts
- Subtle animations and hover effects

## 🚀 Deployment

### Static Build

```bash
npm run build
```

This generates static files in the `build/` directory that can be served by any static hosting service.

### Deployment Options

- **GitHub Pages**: Configure in `docusaurus.config.ts` and use `npm run deploy`
- **Netlify/Vercel**: Connect repository and set build command to `npm run build`
- **Internal Hosting**: Serve the `build/` directory with any web server

### Environment Variables

For production deployments, update these values in `docusaurus.config.ts`:

- `url`: Your production domain
- `baseUrl`: Base path if not serving from root
- `organizationName` and `projectName`: For GitHub Pages deployment

## 🔄 Maintenance

### Keeping Content Fresh

1. **API Changes**: Update documentation when the Go API schema changes
2. **Swagger Sync**: Reference `../goapi/docs/swagger.yaml` for schema accuracy
3. **Examples**: Verify code examples work with current API version
4. **Links**: Check external links periodically

### Performance

- Images are automatically optimized by Docusaurus
- The site generates a service worker for offline functionality
- Bundle analysis available with `npm run build -- --bundle-analyzer`

## 🤝 Contributing

1. Make changes to documentation files
2. Test locally with `npm start`
3. Build to verify no broken links: `npm run build`
4. Submit pull request with clear description of changes

## 📚 Resources

- [Docusaurus Documentation](https://docusaurus.io/docs)
- [Markdown Guide](https://www.markdownguide.org/)
- [MDX Documentation](https://mdxjs.com/) (for advanced components)