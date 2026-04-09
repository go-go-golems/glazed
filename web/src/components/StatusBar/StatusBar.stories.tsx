import type { Meta, StoryObj } from '@storybook/react';
import { StatusBar } from './StatusBar';

const meta: Meta<typeof StatusBar> = {
  title: 'Components/StatusBar',
  component: StatusBar,
  tags: ['autodocs'],
};

export default meta;

export const Default: StoryObj = { args: { count: 42 } };
export const OneSection: StoryObj = { args: { count: 1 } };
export const ZeroSections: StoryObj = { args: { count: 0 } };
