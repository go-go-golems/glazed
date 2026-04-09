import type { Meta, StoryObj } from '@storybook/react';
import { EmptyState } from './EmptyState';

const meta: Meta<typeof EmptyState> = {
  title: 'Components/EmptyState',
  component: EmptyState,
  tags: ['autodocs'],
  decorators: [(Story) => (
    <div style={{ display: 'flex', flex: 1, alignItems: 'center', justifyContent: 'center', background: '#fff' }}>
      <Story />
    </div>
  )],
};

export default meta;

export const Default: StoryObj = { args: {} };
export const CustomLabel: StoryObj = { args: { label: 'Choose a section to read.' } };
