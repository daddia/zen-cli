# Component View

## Overview

The Component view (C4 Level 3) shows the internal structure of key containers, their components, and how they interact. This view focuses on the main architectural components within the Zen CLI system.

## Core Components Architecture

```mermaid
graph TB
    subgraph "Command Layer Components"
        Factory[Factory<br/>Dependency Injection]
        RootCmd[Root Command<br/>Global Config]
        SubCmds[Subcommands<br/>version, init, config, status]
        IOStreams[IO Streams<br/>Input/Output Abstraction]
    end
    
    subgraph "Core Services Components"
        ConfigLoader[Config Loader<br/>Multi-source Loading]
        Validator[Config Validator<br/>Schema Validation]
        Logger[Logger Interface<br/>Structured Logging]
        ErrorHandler[Error Handler<br/>Categorized Errors]
    end
    
    subgraph "AI Agent Components"
        AgentManager[Agent Manager<br/>Orchestration]
        ProviderFactory[Provider Factory<br/>LLM Selection]
        PromptEngine[Prompt Engine<br/>Template Management]
        ContextManager[Context Manager<br/>Conversation State]
        TokenCounter[Token Counter<br/>Usage Tracking]
    end
    
    Factory --> RootCmd
    Factory --> IOStreams
    Factory --> Logger
    
    RootCmd --> SubCmds
    SubCmds --> AgentManager
    
    ConfigLoader --> Validator
    AgentManager --> ProviderFactory
    AgentManager --> PromptEngine
    AgentManager --> ContextManager
    ProviderFactory --> TokenCounter
```

## Component Details

### Command Layer (`pkg/cmd/`)

#### Factory Component
```go
type Factory interface {
    IOStreams() *iostreams.IOStreams
    Config() (*config.Config, error)
    Logger() logging.Logger
    WorkspaceManager() (WorkspaceManager, error)
    AgentManager() (AgentManager, error)
}
```
- **Pattern**: Abstract Factory
- **Purpose**: Centralized dependency injection
- **Benefits**: Testability, lazy initialization, clean dependencies

#### Command Components
- **Root Command**: Global flags, help system, subcommand routing
- **Version Command**: Build info, dependencies, platform details
- **Init Command**: Workspace setup, configuration initialization
- **Config Command**: Configuration management (get, set, list)
- **Status Command**: System health, integration status

### Core Services (`internal/`)

#### Configuration Components
```mermaid
classDiagram
    class ConfigLoader {
        +Load() Config
        +Validate() error
        -loadFromFile()
        -loadFromEnv()
        -loadFromFlags()
    }
    
    class ConfigValidator {
        +Validate(config) error
        +ValidateField(field) error
        -rules map
    }
    
    class Config {
        +LogLevel string
        +LogFormat string
        +CLI CLIConfig
        +Workspace WorkspaceConfig
    }
    
    ConfigLoader --> Config
    ConfigLoader --> ConfigValidator
```

#### Logging Components
- **Logger Interface**: Abstract logging interface
- **Logrus Implementation**: Concrete implementation
- **Formatters**: Text and JSON output formatters
- **Hooks**: Custom log processing hooks

### AI Agent System (`internal/agents/`)

#### Agent Manager
```mermaid
stateDiagram-v2
    [*] --> Initialize
    Initialize --> SelectProvider
    SelectProvider --> LoadPrompt
    LoadPrompt --> ExecuteRequest
    ExecuteRequest --> ProcessResponse
    ProcessResponse --> FormatOutput
    FormatOutput --> [*]
    
    ExecuteRequest --> HandleError: Error
    HandleError --> RetryLogic
    RetryLogic --> ExecuteRequest: Retry
    RetryLogic --> FormatOutput: Max Retries
```

#### Provider Components
- **Provider Interface**: Common LLM interface
- **OpenAI Provider**: OpenAI API implementation
- **Anthropic Provider**: Claude API implementation  
- **Azure Provider**: Azure OpenAI implementation
- **Local Provider**: Local model support

#### Context Management
```go
type ContextManager struct {
    conversations map[string]*Conversation
    maxTokens     int
    windowSize    int
}

type Conversation struct {
    ID       string
    Messages []Message
    Tokens   int
    Metadata map[string]interface{}
}
```

### Zenflow Engine (`internal/workflow/`)

The Zenflow Engine orchestrates the 7-stage unified workflow that standardizes how teams move from strategy to shipped value. For detailed documentation, see the [Zenflow Guide](../../zen-workflow/).

```mermaid
graph LR
    subgraph "Zenflow Components"
        StateMachine[State Machine<br/>Transition Logic]
        StateStore[State Store<br/>Persistence]
        Executor[Stage Executor<br/>Stage Processing]
        Validator[Quality Gates<br/>Stage Validation]
    end
    
    subgraph "Zenflow Stages"
        S1[1. Align<br/>Define Success]
        S2[2. Discover<br/>Gather Evidence]
        S3[3. Prioritize<br/>Rank by Value]
        S4[4. Design<br/>Specify Solution]
        S5[5. Build<br/>Implement]
        S6[6. Ship<br/>Deploy Safely]
        S7[7. Learn<br/>Measure Outcomes]
    end
    
    StateMachine --> StateStore
    StateMachine --> Executor
    Executor --> Validator
    
    S1 --> S2 --> S3 --> S4 --> S5 --> S6 --> S7
    S7 -.-> S1
```

#### Key Features
- **Cross-functional support**: Product managers, designers, engineers, and analysts use the same workflow
- **Quality gates**: Automated and manual checkpoints ensure standards before progression
- **Workflow streams**: Specialized implementations (I2D, C2M, D2S) for different work types
- **State persistence**: Reliable state storage with crash recovery capabilities

### Integration Components (`internal/integrations/`)

#### Plugin Architecture
```mermaid
classDiagram
    class IntegrationPlugin {
        <<interface>>
        +Name() string
        +Connect() error
        +Execute(action) Result
        +Disconnect() error
    }
    
    class JiraPlugin {
        +Connect() error
        +CreateIssue() Issue
        +UpdateIssue() error
        +QueryIssues() []Issue
    }
    
    class GitHubPlugin {
        +Connect() error
        +CreatePR() PullRequest
        +ListIssues() []Issue
        +CreateRelease() Release
    }
    
    class SlackPlugin {
        +Connect() error
        +SendMessage() error
        +PostToChannel() error
    }
    
    IntegrationPlugin <|-- JiraPlugin
    IntegrationPlugin <|-- GitHubPlugin
    IntegrationPlugin <|-- SlackPlugin
```

### Template Engine (`internal/templates/`)

#### Template Components
- **Template Registry**: Template discovery and management
- **Parser**: Go template parsing with custom functions
- **Executor**: Template execution with context
- **Custom Functions**: Domain-specific template functions

```go
type TemplateEngine struct {
    registry  *TemplateRegistry
    parser    *template.Template
    functions template.FuncMap
}

type Template struct {
    Name     string
    Content  string
    Type     TemplateType
    Metadata map[string]interface{}
}
```

## Component Interactions

### Command Execution Flow
```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Factory
    participant Command
    participant Service
    participant Integration
    
    User->>CLI: Command input
    CLI->>Factory: Create dependencies
    Factory->>Command: Inject dependencies
    Command->>Service: Execute business logic
    Service->>Integration: External calls
    Integration-->>Service: Results
    Service-->>Command: Processed data
    Command-->>User: Formatted output
```

### Error Handling Flow
```mermaid
graph TD
    Error[Error Occurs] --> Categorize[Categorize Error]
    Categorize --> Silent{Silent Error?}
    Silent -->|Yes| NoOutput[No Output]
    Silent -->|No| UserCancel{User Cancel?}
    UserCancel -->|Yes| ExitCancel[Exit Code 2]
    UserCancel -->|No| NoResults{No Results?}
    NoResults -->|Yes| ExitOK[Exit Code 0]
    NoResults -->|No| PrintError[Print Error with Suggestion]
    PrintError --> ExitError[Exit Code 1]
```

## Key Design Decisions

### Dependency Injection
- Factory pattern for clean dependency management
- Lazy initialization for performance
- Interface-based design for testability

### Error Management  
- Categorized error types
- Structured exit codes
- Helpful error suggestions

### Plugin System
- Interface-based plugin contracts
- Dynamic plugin loading
- Isolated plugin execution

### Configuration
- Multi-source configuration with clear precedence
- Schema validation at load time
- Environment-specific profiles
