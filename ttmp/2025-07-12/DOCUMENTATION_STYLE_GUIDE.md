# Glazed Documentation Style Guide

## Overview

This guide establishes consistent writing principles for Glazed documentation to ensure content is professional, clear, and accessible to developers without unnecessary conversational elements.

## Core Principles

### 1. Professional Technical Writing
- **Write for developers**: Use appropriate technical language without oversimplification
- **Be direct and factual**: Explain concepts clearly without conversational filler
- **Provide context and purpose**: Start sections with clear objectives and use cases
- **Avoid analogies and cute language**: Focus on technical accuracy over entertainment

### 2. Clear Information Architecture
- **Start with overview, then detail**: Begin with high-level concepts before implementation specifics
- **Use descriptive section hierarchy**: Organize content for both linear reading and quick reference
- **Enable quick scanning**: Use bullet points, code blocks, and clear headings
- **Include cross-references**: Link to related concepts and implementations

### 3. Practical Implementation Focus
- **Include working examples**: Every concept should have concrete, runnable code
- **Show implementation patterns**: Demonstrate real-world usage patterns
- **Connect to specific use cases**: Explain when and why to use each feature
- **Address common issues**: Include troubleshooting and best practices

## Content Types and Structure

### Topics (Reference Materials)
**Purpose**: Comprehensive coverage of a system or concept
**Structure**:
```
1. Overview and purpose
2. Architecture and key concepts
3. Implementation details and patterns
4. Advanced usage and best practices
5. Common issues and solutions
```
**Tone**: Professional and authoritative technical reference

### Tutorials (Learning Experiences)  
**Purpose**: Hands-on learning through practical implementation
**Structure**:
```
1. Learning objectives and prerequisites
2. Setup and initial implementation
3. Step-by-step development with code examples
4. Key concepts and explanations
5. Next steps and related topics
```
**Tone**: Clear instructional guidance without excessive encouragement

### Applications (Complete Examples)
**Purpose**: Demonstrate real-world implementation patterns
**Structure**:
```
1. Requirements and problem description
2. Solution architecture and approach
3. Complete implementation walkthrough
4. Alternative approaches and variations
5. Extension possibilities
```
**Tone**: Practical and implementation-focused

### Reference (API Documentation)
**Purpose**: Precise technical specifications and usage
**Structure**:
```
1. Purpose and overview
2. API signatures and parameters
3. Usage examples and patterns
4. Return values, errors, and edge cases
5. Related functions and cross-references
```
**Tone**: Precise and technically accurate with practical guidance

## Writing Guidelines

### Language and Voice
- **Use active voice**: "The command processes data" not "Data is processed by the command"
- **Write directly**: Use third person for technical descriptions, second person for instructions
- **Use present tense**: "This function returns..." not "This function will return..."
- **Be specific and precise**: Instead of "various options," list actual options
- **Define technical terms**: Explain domain-specific concepts clearly

### Code Examples
- **Provide complete, runnable examples**: Include all necessary imports and setup
- **Use meaningful examples**: Choose realistic scenarios over generic "foo/bar" examples
- **Comment implementation details**: Explain non-obvious logic and patterns
- **Show expected outcomes**: Include relevant input/output examples
- **Structure progressively**: Start with basic examples, then show advanced patterns

### Content Structure
- **Start with clear objectives**: Define what the section covers and its purpose
- **Use descriptive, scannable headings**: Make content structure immediately clear
- **Organize logically**: Present information in order of increasing complexity
- **Provide clear next steps**: Link to related concepts and follow-up actions
- **Include comprehensive cross-references**: Help readers navigate between related topics

### Content Presentation
- **Use diagrams for architecture**: ASCII art or mermaid diagrams for complex systems
- **Show command output**: Include examples of successful execution
- **Format code appropriately**: Distinguish between CLI commands, code, and output
- **Structure lists clearly**: Use appropriate formatting for different types of information

## Content Quality Checklist

### Before Publishing
- [ ] Does the introduction clearly define the purpose and scope?
- [ ] Are all code examples tested and complete?
- [ ] Is content organized logically with clear progression?
- [ ] Are technical terms and concepts explained clearly?
- [ ] Do headings accurately describe their content?
- [ ] Are cross-references accurate and helpful?
- [ ] Is the tone professional and appropriate for developers?
- [ ] Can developers understand and apply the information effectively?

### Review Criteria
- **Clarity**: Is the technical content clear and unambiguous?
- **Completeness**: Are there gaps in coverage or missing implementation details?
- **Accuracy**: Are all examples, APIs, and references correct?
- **Practicality**: Does this help developers accomplish their implementation goals?
- **Consistency**: Does this follow our established style and structure guidelines?

## Common Patterns

### Introducing Technical Concepts
```markdown
## Parameter Layers

Parameter layers organize related command parameters into reusable groups. This approach addresses several common CLI development challenges:

- **Parameter organization**: Group related flags logically (database, logging, output)
- **Code reuse**: Share parameter definitions across multiple commands  
- **Namespace management**: Avoid flag name conflicts in complex applications
- **Configuration sources**: Support CLI flags, environment variables, and config files consistently

### Architecture

Parameter layers work by collecting parameter definitions at design time and resolving values from multiple sources at runtime.

[Architecture diagram]

### Implementation

The basic pattern involves defining parameter layers and composing them into commands:

```go
// Define a reusable layer
databaseLayer := schema.NewSection("database", "Database connection parameters",
    fields.New("host", fields.TypeString,
        fields.WithDefault("localhost")),
    fields.New("port", fields.TypeInteger,
        fields.WithDefault(5432)),
)

// Use in command definition
cmd := cmds.NewCommandDescription("mycommand",
    cmds.WithLayersList(databaseLayer, loggingLayer),
)
```

### Usage in Commands

Commands extract typed values from parsed layers:

```go
func (c *MyCommand) Run(ctx context.Context, parsedLayers *values.Values) error {
    settings := &DatabaseSettings{}
    if err := parsedLayers.InitializeStruct("database", settings); err != nil {
        return err
    }
    
    // Use settings.Host, settings.Port, etc.
    return nil
}
```

This pattern provides type safety and automatic validation while maintaining clean separation between parameter definition and business logic.
```

This style guide ensures documentation is technically accurate, professionally presented, and practically useful for developers implementing Glazed-based applications.
