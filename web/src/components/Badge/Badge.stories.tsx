import type { Meta, StoryObj } from '@storybook/react';
import { Badge } from './Badge';

const meta: Meta<typeof Badge> = {
  title: 'Components/Badge',
  component: Badge,
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['type', 'topic', 'command', 'flag'],
    },
  },
};

export default meta;
type Story = StoryObj<typeof Badge>;

export const Topic: Story = { args: { text: 'GeneralTopic', variant: 'type' } };
export const Example: Story = { args: { text: 'Example', variant: 'type' } };
export const App: Story = { args: { text: 'Application', variant: 'type' } };
export const Tutorial: Story = { args: { text: 'Tutorial', variant: 'type' } };
export const TopicBadge: Story = { args: { text: 'database', variant: 'topic' } };
export const CommandBadge: Story = { args: { text: 'glaze', variant: 'command' } };
export const FlagBadge: Story = { args: { text: '--verbose', variant: 'flag' } };
