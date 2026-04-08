import type { Meta, StoryObj } from '@storybook/react';
import { TitleBar } from './TitleBar';

const meta: Meta<typeof TitleBar> = {
  title: 'Components/TitleBar',
  component: TitleBar,
  tags: ['autodocs'],
};

export default meta;

export const Default: StoryObj = { args: { title: 'Glazed Help Browser' } };
export const SectionsWindow: StoryObj = { args: { title: '📁 Sections' } };
export const ContentWindow: StoryObj = { args: { title: '📄 Introduction — glaze help intro' } };
