import type { Meta, StoryObj } from '@storybook/react';
import { SectionList } from './SectionList';
import type { SectionSummary } from '../../types';

const meta: Meta<typeof SectionList> = {
  title: 'Components/SectionList',
  component: SectionList,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof SectionList>;

const SECTIONS: SectionSummary[] = [
  { id: 1, slug: 'intro', type: 'GeneralTopic', title: 'Introduction', short: 'Welcome to Glazed.', topics: ['help'], isTopLevel: true },
  { id: 2, slug: 'database', type: 'Example', title: 'Database', short: 'How to connect to a database.', topics: ['database'], isTopLevel: false },
  { id: 3, slug: 'config', type: 'GeneralTopic', title: 'Configuration', short: 'Configure the application.', topics: ['config'], isTopLevel: false },
];

export const WithItems: Story = {
  args: { sections: SECTIONS, activeSlug: null, onSelect: () => {} },
};

export const WithActiveItem: Story = {
  args: { sections: SECTIONS, activeSlug: 'database', onSelect: () => {} },
};

export const Empty: Story = {
  args: { sections: [], activeSlug: null, onSelect: () => {} },
};
