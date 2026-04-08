import type { Meta, StoryObj } from '@storybook/react';
import { useState } from 'react';
import { TypeFilter, type FilterValue } from './TypeFilter';

const meta: Meta<typeof TypeFilter> = {
  title: 'Components/TypeFilter',
  component: TypeFilter,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof TypeFilter>;

export const AllActive: Story = { args: { value: 'All', onChange: () => {} } };
export const TopicActive: Story = { args: { value: 'GeneralTopic', onChange: () => {} } };
export const ExampleActive: Story = { args: { value: 'Example', onChange: () => {} } };

export const Interactive: Story = {
  render: () => {
    const [value, setValue] = useState<FilterValue>('All');
    return <TypeFilter value={value} onChange={setValue} />;
  },
};
