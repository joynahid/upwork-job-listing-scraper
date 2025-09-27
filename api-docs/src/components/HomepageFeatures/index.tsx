import type {ReactNode} from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import type {LucideIcon} from 'lucide-react';
import {NotebookPen, Send, Signal} from 'lucide-react';

import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  description: string;
  icon: LucideIcon;
};

const features: FeatureItem[] = [
  {
    title: 'Newsletter-ready collections',
    description:
      'Cluster briefs by theme, budget, and buyer intent so your editorial calendar fills itself with credible story ideas.',
    icon: NotebookPen,
  },
  {
    title: 'Audience intelligence in every call',
    description:
      'Buyer spend, hiring velocity, and required skills arrive with each record, perfect for framing angles and CTAs that resonate.',
    icon: Signal,
  },
  {
    title: 'Frictionless distribution',
    description:
      'Send curated jobs to Discord, Telegram, and inbox digests as soon as matches land, keeping communities and clients engaged.',
    icon: Send,
  },
];

function Feature({title, description, icon: Icon}: FeatureItem) {
  return (
    <div className={clsx('col col--4', styles.card)}>
      <div className={styles.cardIconWrap}>
        <Icon aria-hidden="true" className={styles.cardIcon} />
      </div>
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
          {features.map((feature) => (
            <Feature key={feature.title} {...feature} />
          ))}
        </div>
      </div>
    </section>
  );
}
