import type { Meta, StoryObj } from '@storybook/react';
import { SectionCard } from './SectionCard';
import type { SectionSummary } from '../../types';

const meta: Meta<typeof SectionCard> = {
  title: 'Components/SectionCard',
  component: SectionCard,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof SectionCard>;

const exampleSection: SectionSummary = {
  id: 1,
  slug: 'database',
  type: 'Example',
  title: 'Database Example',
  short: 'How to connect to a database with connection pooling.',
  topics: ['database', 'sql'],
};

export const Inactive: Story = {
  args: { section: exampleSection, isActive: false, onClick: () => {} },
};

export const Active: Story = {
  args: { section: exampleSection, isActive: true, onClick: () => {} },
};

export const TopLevel: Story = {
  args: {
    section: { ...exampleSection, topics: ['getting-started'] },
    isActive: false,
    onClick: () => {},
  },
};
