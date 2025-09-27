import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  apiSidebar: [
    'overview',
    'use-cases',
    'pricing',
    'getting-started',
    {
      type: 'category',
      label: 'API Documentation',
      collapsed: false,
      items: [
        'api/authentication',
        'api/endpoints',
        'api/filtering',
      ],
    },
    {
      type: 'category',
      label: 'Support',
      collapsed: false,
      items: [
        'support/faq',
        'support/rate-limits',
        'support/contact',
      ],
    },
  ],
};

export default sidebars;
