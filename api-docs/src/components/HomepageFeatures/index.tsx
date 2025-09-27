import type {ReactNode} from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

const FeatureList = [
  {
    title: 'Real-Time Upwork Job Data',
    description:
      'Access fresh, high-quality Upwork job postings with verified client information, budgets, and hiring activity. Perfect for lead generation and market analysis.',
  },
  {
    title: 'Advanced Filtering & Search',
    description:
      'Find exactly what you need with powerful filters for budget ranges, client verification status, skills, location, and posting dates. Save time with precise targeting.',
  },
  {
    title: 'Enterprise-Grade Reliability',
    description:
      'Built for scale with 99.9% uptime, secure API authentication, and comprehensive documentation. Trusted by agencies, freelancers, and businesses worldwide.',
  },
] as const;

function Feature({title, description}: (typeof FeatureList)[number]) {
  return (
    <div className={clsx('col col--4', styles.card)}>
      <Heading as="h3" className={styles.cardTitle}>
        {title}
      </Heading>
      <p className={styles.cardBody}>{description}</p>
    </div>
  );
}

export default function HomepageFeatures(): ReactNode {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((feature) => (
            <Feature key={feature.title} {...feature} />
          ))}
        </div>
      </div>
    </section>
  );
}
