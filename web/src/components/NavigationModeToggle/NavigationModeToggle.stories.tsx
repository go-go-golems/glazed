import type { Meta, StoryObj } from '@storybook/react';
import { NavigationModeToggle } from './NavigationModeToggle';

const meta: Meta<typeof NavigationModeToggle> = {
  title: 'Components/NavigationModeToggle',
  component: NavigationModeToggle,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof NavigationModeToggle>;

export const TreeSelected: Story = {
  args: {
    value: 'tree',
    onChange: () => {},
  },
};

export const SearchSelected: Story = {
  args: {
    value: 'search',
    onChange: () => {},
  },
};
