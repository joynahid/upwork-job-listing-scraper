import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'Upwork Jobs API',
  tagline: 'Access premium Upwork job data through our powerful API service.',
  favicon: 'img/favicon.ico',
  future: {
    v4: true,
  },
  url: 'https://docs.upworkjobapi.local',
  baseUrl: '/',
  organizationName: 'upwork-automation',
  projectName: 'upworkjobposting',
  onBrokenLinks: 'warn',
  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },
  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          routeBasePath: '/docs',
          editUrl: undefined,
          showLastUpdateTime: true,
          showLastUpdateAuthor: false,
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],
  themeConfig: {
    image: 'img/docusaurus-social-card.jpg',
    colorMode: {
      respectPrefersColorScheme: true,
      defaultMode: 'dark',
    },
    navbar: {
      title: 'Upwork Jobs API',
      logo: {
        alt: 'Upwork Job API Logo',
        src: 'img/logo.svg',
      },
      items: [
        {to: '/docs', label: 'Documentation', position: 'left'},
        {to: '/docs/pricing', label: 'Pricing', position: 'left'},
        {to: '/docs/use-cases', label: 'Use Cases', position: 'left'},
        {
          href: 'mailto:sales@upworkjobsapi.com',
          label: 'Contact Sales',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Product',
          items: [
            {
              label: 'API Documentation',
              to: '/docs',
            },
            {
              label: 'Pricing Plans',
              to: '/docs/pricing',
            },
            {
              label: 'Use Cases',
              to: '/docs/use-cases',
            },
          ],
        },
        {
          title: 'Support',
          items: [
            {
              label: 'Contact Sales',
              href: 'mailto:sales@upworkjobsapi.com',
            },
            {
              label: 'Technical Support',
              href: 'mailto:support@upworkjobsapi.com',
            },
            {
              label: 'Status Page',
              href: 'https://status.upworkjobsapi.com',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Upwork Jobs API. All rights reserved.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
    },
    docs: {
      sidebar: {
        hideable: true,
      },
    },
    tableOfContents: {
      maxHeadingLevel: 4,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
