import type { Meta, StoryObj } from '@storybook/react';
import { useState } from 'react';
import { SearchBar } from './SearchBar';

const meta: Meta<typeof SearchBar> = {
  title: 'Components/SearchBar',
  component: SearchBar,
  tags: ['autodocs'],
  decorators: [(Story) => <div style={{ width: 240 }}><Story /></div>],
};

export default meta;
type Story = StoryObj<typeof SearchBar>;

export const Empty: Story = { args: { value: '', onChange: () => {} } };

export const WithText: Story = {
  args: { value: 'database', onChange: () => {} },
};

export const Interactive: Story = {
  render: () => {
    const [value, setValue] = useState('');
    return <SearchBar value={value} onChange={setValue} />;
  },
};
