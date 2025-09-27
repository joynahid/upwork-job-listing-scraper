import type {ReactNode} from 'react';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import HomepageFeatures from '@site/src/components/HomepageFeatures';
import Heading from '@theme/Heading';
import type {LucideIcon} from 'lucide-react';
import {BarChart3, MessageCircle, Share2, Sparkles, Workflow} from 'lucide-react';

import styles from './index.module.css';

type Highlight = {
  icon: LucideIcon;
  title: string;
  description: string;
};

const highlights: Highlight[] = [
  {
    icon: Sparkles,
    title: 'Creator-grade briefs',
    description: 'Summaries, budgets, and buyer context extracted from every Upwork brief so you can pitch angles fast.',
  },
  {
    icon: BarChart3,
    title: 'Signal-rich insights',
    description: 'Trusted metadata on spend, hiring velocity, and skills to map your next newsletter segment or idea board.',
  },
  {
    icon: Workflow,
    title: 'Automation ready',
    description: 'Push new jobs into your n8n, Zapier, or Make pipelines, draft talking points, and deliver updates automatically.',
  },
];

function HeroHighlight({icon: Icon, title, description}: Highlight): ReactNode {
  return (
    <div className={styles.highlight}>
      <Icon aria-hidden="true" className={styles.highlightIcon} />
      <div>
        <p className={styles.highlightTitle}>{title}</p>
        <p className={styles.highlightBody}>{description}</p>
      </div>
    </div>
  );
}

function IntegrationCallout(): ReactNode {
  return (
    <section className={styles.integrationSection}>
      <div className="container">
        <div className={styles.integrationGrid}>
          <div className={styles.integrationCard}>
            <div className={styles.integrationIconWrapper}>
              <Share2 aria-hidden="true" className={styles.integrationIcon} />
            </div>
            <h3>Workflow integrations</h3>
            <p>
              Ship cleaned job data straight into n8n, Zapier, and Make (Integromat) with ready-to-use webhook templates. Trigger
              automations the instant a brief matches your saved filters.
            </p>
            <ul>
              <li>Map fields to your idea capture or CRM tables</li>
              <li>Auto-generate briefs and topics for newsletters</li>
              <li>Route approvals to Notion, Airtable, or Google Sheets</li>
            </ul>
          </div>
          <div className={styles.integrationCard}>
            <div className={styles.integrationIconWrapper}>
              <MessageCircle aria-hidden="true" className={styles.integrationIcon} />
            </div>
            <h3>Audience channels</h3>
            <p>
              Keep collaborators and subscribers in the loop with native Discord and Telegram delivery. Share curated deal flow and
              article-ready prompts where your community already lives.
            </p>
            <ul>
              <li>Stream highlights into Discord announcement channels</li>
              <li>Send Telegram digests with your top angles</li>
              <li>Hand off winning leads to teammates in real time</li>
            </ul>
          </div>
        </div>
      </div>
    </section>
  );
}

const schemaSnippet = `{
  "success": true,
  "data": [
    {
      "id": "upwork-872341",
      "title": "Launch a weekly AI founder newsletter",
      "description": "We need a researcher-writer to source stories and trends...",
      "posted_on": "2024-10-24T08:12:43Z",
      "category": {
        "name": "Writing & Translation",
        "group": "Sales & Marketing"
      },
      "budget": {
        "fixed_amount": 2500,
        "currency": "USD"
      },
      "buyer": {
        "payment_verified": true,
        "country": "US",
        "total_spent": 84500,
        "last_activity": "2024-10-24T07:55:12Z"
      },
      "skills": ["newsletter", "ai research", "marketing"],
      "tags": ["founder stories", "growth marketing"],
      "client_activity": {
        "total_applicants": 12,
        "total_hired": 4
      }
    }
  ],
  "count": 1,
  "last_updated": "2024-10-24T08:14:03Z"
}`;

function SchemaPreview(): ReactNode {
  return (
    <section className={styles.schemaSection}>
      <div className="container">
        <div className={styles.schemaGrid}>
          <div>
            <h2>Job schema at a glance</h2>
            <p>
              Every record arrives pre-ranked for creators: consistent identifiers, market signals, and metadata you can plug into idea
              prompts, editorial calendars, and pitch trackers without additional cleanup.
            </p>
            <ul>
              <li>Stable IDs for versioning and deduplication</li>
              <li>Buyer context to qualify briefs before you draft</li>
              <li>Skills and tags ready for clustering or topic models</li>
            </ul>
            <div className={styles.schemaCta}>
              <Link className="button button--primary button--lg" to="/docs/api/endpoints">
                Explore API endpoints
              </Link>
              <Link className="button button--outline button--lg" to="/docs/getting-started">
                Build your first automation
              </Link>
            </div>
          </div>
          <pre className={styles.schemaPreview}>
            <code>{schemaSnippet}</code>
          </pre>
        </div>
      </div>
    </section>
  );
}

function HomepageHeader(): ReactNode {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header className="hero hero--primary">
      <div className={`container ${styles.heroBanner}`}>
        <div className={styles.heroContent}>
          <Heading as="h1" className="hero__title">
            Turn live Upwork demand into daily content ideas
          </Heading>
          <p className="hero__subtitle">
            {siteConfig.tagline} Curated for creators building newsletters, idea banks, and growth content. Set filters once, get
            ready-to-ship briefs forever.
          </p>
          <div className={styles.buttons}>
            <Link className="button button--secondary button--lg" to="/docs/getting-started">
              Start free trial
            </Link>
            <Link className="button button--outline button--lg" to="/docs">
              See how it works
            </Link>
          </div>
          <div className={styles.heroHighlights}>
            {highlights.map((highlight) => (
              <HeroHighlight key={highlight.title} {...highlight} />
            ))}
          </div>
        </div>
      </div>
    </header>
  );
}

export default function Home(): ReactNode {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout title={siteConfig.title} description={siteConfig.tagline}>
      <HomepageHeader />
      <main>
        <HomepageFeatures />
        <IntegrationCallout />
        <SchemaPreview />
      </main>
    </Layout>
  );
}
