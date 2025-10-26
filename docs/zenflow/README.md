# Zenflow Guide

Zenflow is a comprehensive seven-stage workflow that unifies product development from strategic planning to production deployment and learning. This guide helps you understand and implement Zenflow in your organization.

## What is Zenflow?

Zenflow provides a structured approach to product development that:

- **Unifies teams** - Product managers, designers, engineers, and analysts use the same workflow
- **Automates processes** - CLI commands orchestrate tools and enforce quality standards
- **Measures outcomes** - Every stage produces measurable value indicators
- **Reduces complexity** - Seven clear stages replace fragmented, tool-specific workflows

## Quick Start

Zenflow follows seven sequential stages, each with clear objectives:

1. **[Align](stages.md#align)** - Define what success looks like
2. **[Discover](stages.md#discover)** - Gather evidence and insights
3. **[Prioritize](stages.md#prioritize)** - Rank work by value and effort
4. **[Design](stages.md#design)** - Specify what you'll build
5. **[Build](stages.md#build)** - Implement with quality
6. **[Ship](stages.md#ship)** - Deploy safely to production
7. **[Learn](stages.md#learn)** - Measure outcomes and iterate

## Core Concepts

### The Seven-Stage Workflow

Each stage has three key elements:

- **Goal** - What you're trying to accomplish
- **Artifact** - What you produce as evidence of completion
- **Quality Gate** - Criteria that must be met to proceed

### Quality Gates

[Quality gates](quality-gates.md) ensure consistent standards throughout development:

- Automated checks validate code quality, security, and performance
- Manual reviews confirm business alignment and user experience
- Progressive rigor increases as work approaches production

### Workflow Streams

[Streams](streams.md) are specialized implementations for different work types:

- **I2D (Idea to Delivery)** - Product discovery and validation
- **C2M (Code to Market)** - Engineering implementation
- **D2S (Deploy to Scale)** - Production deployment and operations

## Getting Started

### Prerequisites

Before starting with Zenflow:

1. Install the Zen CLI (see [Installation Guide](../installation/))
2. Configure workspace settings
3. Connect to your development tools (Jira, GitHub, etc.)

### Your First Zenflow Cycle

Follow our [Getting Started Guide](getting-started.md) to run your first complete workflow cycle:

1. Create a new initiative with `zen align`
2. Research requirements with `zen discover`
3. Prioritize features with `zen prioritize`
4. Design specifications with `zen design`
5. Build implementation with `zen build`
6. Deploy to production with `zen ship`
7. Measure results with `zen learn`

## Documentation Structure

- **[Getting Started](getting-started.md)** - Step-by-step guide for first-time users
- **[Stages](stages.md)** - Detailed documentation for each workflow stage
- **[Commands](commands.md)** - CLI command reference and examples
- **[Quality Gates](quality-gates.md)** - Quality standards and enforcement
- **[Streams](streams.md)** - Specialized workflow implementations
- **[Best Practices](best-practices.md)** - Tips and recommended patterns
- **[Troubleshooting](troubleshooting.md)** - Common issues and solutions

## Key Benefits

### For Product Managers

- Clear progression from strategy to outcomes
- Automated stakeholder alignment and documentation
- Data-driven prioritization and decision making
- Direct connection between initiatives and metrics

### For Designers

- Integrated design process within development workflow
- Design system compliance and accessibility validation
- User research and validation built into stages
- Clear handoffs to engineering teams

### For Engineers

- Contracts-first development with code generation
- Automated quality checks and testing
- Progressive deployment with rollback capabilities
- Focus on implementation rather than process

### For Analytics Teams

- Measurement framework integrated throughout workflow
- Automated experiment design and validation
- Clear success criteria and outcome tracking
- Data-driven insights feed next iterations

## Integration with Existing Tools

Zenflow works with your existing tools rather than replacing them:

- **Work Management**: Jira, Linear, Asana
- **Design**: Figma, Sketch, Adobe XD
- **Development**: GitHub, GitLab, Bitbucket
- **CI/CD**: Jenkins, CircleCI, GitHub Actions
- **Analytics**: Google Analytics, Mixpanel, Amplitude
- **Monitoring**: DataDog, New Relic, Prometheus

## Support and Resources

- **[Command Reference](commands.md)** - Complete CLI documentation
- **[API Documentation](../api/)** - Programmatic access to Zenflow
- **[Contributing Guide](../contributing/)** - Help improve Zenflow
- **[Community Forum](https://community.zen.dev)** - Get help and share experiences

## Next Steps

1. Complete the [Getting Started Guide](getting-started.md)
2. Review [Stage Documentation](stages.md) for your role
3. Explore [Best Practices](best-practices.md) from successful teams
4. Join the [Zenflow Community](https://community.zen.dev)
