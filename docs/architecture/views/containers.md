# Container View

## Overview

The Container view (C4 Level 2) shows the high-level shape of the Zen CLI system architecture and how responsibilities are distributed across containers. A container represents an application or data store.

## Container Diagram

```mermaid
graph TB
    subgraph "Zen CLI System"
        subgraph "Application Layer"
            CLI[CLI Interface<br/>Cobra Commands]
            Orchestrator[Command Orchestrator<br/>Error Handling & Exit Codes]
        end
        
        subgraph "Core Services"
            Config[Configuration Service<br/>Viper-based Multi-source]
            Logger[Logging Service<br/>Structured Logging]
            Workspace[Workspace Manager<br/>Project Context]
        end
        
        subgraph "Business Logic"
            Agents[AI Agent System<br/>LLM Orchestration]
            Workflow[Workflow Engine<br/>State Management]
            Templates[Template Engine<br/>Content Generation]
            Quality[Quality Gates<br/>Validation & Testing]
        end
        
        subgraph "Integration Layer"
            Providers[LLM Providers<br/>OpenAI, Anthropic, Azure]
            External[External Integrations<br/>Jira, GitHub, Slack]
            MCP[MCP Server Client<br/>Model Context Protocol]
        end
        
        subgraph "Storage"
            FileSystem[File System<br/>Local Storage]
            Cache[Cache Layer<br/>Response Caching]
        end
    end
    
    CLI --> Orchestrator
    Orchestrator --> Config
    Orchestrator --> Logger
    Orchestrator --> Workspace
    
    CLI --> Agents
    CLI --> Workflow
    CLI --> Templates
    CLI --> Quality
    
    Agents --> Providers
    Workflow --> External
    Agents --> MCP
    
    Workspace --> FileSystem
    Providers --> Cache
```

## Container Responsibilities

### Application Layer

#### CLI Interface
- **Technology**: Go, Cobra Framework
- **Responsibility**: Command parsing, flag handling, user interaction
- **Key Components**: Root command, subcommands, flag definitions

#### Command Orchestrator
- **Technology**: Go
- **Responsibility**: Error handling, exit codes, signal handling
- **Location**: `internal/zencmd`

### Core Services

#### Configuration Service
- **Technology**: Viper
- **Responsibility**: Multi-source configuration management
- **Sources**: CLI flags > Environment > Files > Defaults
- **Location**: `internal/config`

#### Logging Service
- **Technology**: Logrus
- **Responsibility**: Structured logging with multiple outputs
- **Formats**: Text, JSON
- **Location**: `internal/logging`

#### Workspace Manager
- **Responsibility**: Project context, file operations, workspace state
- **Location**: `internal/workspace`

### Business Logic

#### AI Agent System
- **Responsibility**: LLM orchestration, prompt management, context handling
- **Key Features**: Multi-provider support, token management, cost tracking
- **Location**: `internal/agents`

#### Workflow Engine
- **Responsibility**: 12-stage engineering workflow state management
- **Pattern**: State machine with persistent storage
- **Location**: `internal/workflow`

#### Template Engine
- **Technology**: Go Templates
- **Responsibility**: Dynamic content generation for various outputs
- **Location**: `internal/templates`

#### Quality Gates
- **Responsibility**: Code quality, security scanning, test validation
- **Location**: `internal/quality`

### Integration Layer

#### LLM Providers
- **Supported**: OpenAI, Anthropic, Azure OpenAI, Local models
- **Pattern**: Strategy pattern with unified interface
- **Location**: `internal/agents/providers`

#### External Integrations
- **Categories**: Project Management, Version Control, Communication
- **Pattern**: Plugin-based architecture
- **Location**: `internal/integrations`

#### MCP Server Client
- **Responsibility**: Model Context Protocol for AI tool access
- **Location**: `internal/mcp`

### Storage

#### File System
- **Responsibility**: Local file operations, workspace persistence
- **Security**: Path validation, permission checks

#### Cache Layer
- **Responsibility**: Response caching, performance optimization
- **Types**: In-memory, file-based

## Inter-Container Communication

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Orchestrator
    participant Config
    participant Agents
    participant Provider
    
    User->>CLI: zen workflow build
    CLI->>Orchestrator: Execute command
    Orchestrator->>Config: Load configuration
    Config-->>Orchestrator: Configuration
    Orchestrator->>Agents: Process workflow
    Agents->>Provider: Generate content
    Provider-->>Agents: AI response
    Agents-->>CLI: Formatted output
    CLI-->>User: Display result
```

## Technology Stack Summary

| Container | Technology | Purpose |
|-----------|------------|---------|
| CLI Interface | Cobra | Command structure |
| Configuration | Viper | Config management |
| Logging | Logrus | Structured logs |
| AI Agents | Custom + SDKs | LLM orchestration |
| Workflow | State Machine | Workflow management |
| Templates | Go Templates | Content generation |
| Storage | File System | Local persistence |
| Cache | In-memory | Performance |

## Deployment Model

```mermaid
graph LR
    subgraph "Single Binary"
        Binary[zen]
        Embedded[Embedded Assets<br/>Templates, Schemas]
    end
    
    subgraph "Runtime Dependencies"
        Config[Config Files<br/>Optional]
        Workspace[Workspace Dir]
        Env[Environment Vars]
    end
    
    Binary --> Config
    Binary --> Workspace
    Binary --> Env
```

The entire Zen CLI is distributed as a single Go binary with all dependencies compiled in. This enables:
- Zero runtime dependencies
- Simple installation and distribution
- Cross-platform compatibility
- Embedded templates and resources
