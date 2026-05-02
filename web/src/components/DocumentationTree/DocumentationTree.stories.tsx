import type { Meta, StoryObj } from '@storybook/react';
import { DocumentationTree } from './DocumentationTree';
import type { SectionSummary } from '../../types';

const meta: Meta<typeof DocumentationTree> = {
  title: 'Components/DocumentationTree',
  component: DocumentationTree,
  tags: ['autodocs'],
};

export default meta;
type Story = StoryObj<typeof DocumentationTree>;

const SECTIONS: SectionSummary[] = [
  {
    id: 1,
    slug: 'query-dsl',
    type: 'GeneralTopic',
    title: 'Complete User Query DSL Reference',
    short: 'Reference for the query DSL.',
    topics: ['query'],
    isTopLevel: true,
    headings: [
      { id: 'table-of-contents', level: 2, text: 'Table of Contents' },
      { id: 'introduction', level: 2, text: 'Introduction' },
      { id: 'basic-syntax', level: 2, text: 'Basic Syntax' },
      { id: 'field-value-syntax', level: 3, text: 'Field:Value Syntax' },
      { id: 'boolean-operations', level: 3, text: 'Boolean Operations' },
      { id: 'and-operations', level: 4, text: 'AND Operations' },
      { id: 'or-operations', level: 4, text: 'OR Operations' },
      { id: 'finding-examples-and-tutorials', level: 2, text: 'Finding Examples and Tutorials' },
    ],
  },
  {
    id: 2,
    slug: 'adding-fields',
    type: 'GeneralTopic',
    title: 'Adding New Field Types to Glazed',
    short: 'Extend fields.',
    topics: ['fields'],
    isTopLevel: true,
    headings: [],
  },
  {
    id: 3,
    slug: 'examples',
    type: 'Example',
    title: 'Query Examples',
    short: 'Examples for common queries.',
    topics: ['query'],
    isTopLevel: false,
    headings: [{ id: 'simple-shortcuts', level: 2, text: 'Simple Shortcuts' }],
  },
  {
    id: 4,
    slug: 'tutorial',
    type: 'Tutorial',
    title: 'Developer Guide - Using the Query System',
    short: 'Tutorial for developers.',
    topics: ['developer'],
    isTopLevel: false,
    headings: [{ id: 'troubleshooting', level: 2, text: 'Troubleshooting Queries' }],
  },
];

export const TreeWithSubsections: Story = {
  args: {
    sections: SECTIONS,
    search: '',
    activeSlug: 'query-dsl',
    activeHeadingId: 'basic-syntax',
    onSelectDocument: () => {},
    onSelectHeading: () => {},
  },
};

export const FilteredBySubsection: Story = {
  args: {
    sections: SECTIONS,
    search: 'boolean',
    activeSlug: null,
    onSelectDocument: () => {},
    onSelectHeading: () => {},
  },
};

export const Empty: Story = {
  args: {
    sections: [],
    search: '',
    activeSlug: null,
    onSelectDocument: () => {},
    onSelectHeading: () => {},
  },
};
