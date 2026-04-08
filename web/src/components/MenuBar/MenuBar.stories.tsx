import type { Meta, StoryObj } from '@storybook/react';
import { MenuBar } from './MenuBar';

const meta: Meta<typeof MenuBar> = {
  title: 'Components/MenuBar',
  component: MenuBar,
  tags: ['autodocs'],
};

export default meta;

export const Default: StoryObj = { args: {} };
export const CustomTitle: StoryObj = { args: { title: 'Glazed Documentation' } };
