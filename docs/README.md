# Zen CLI Documentation

Welcome to the comprehensive documentation for the Zen AI-powered product lifecycle productivity platform.

## 📚 **Documentation Structure**

### 🎯 **Getting Started**
- [**Project Overview**](overview.md) - Vision, mission, and product capabilities
- [**CLI Structure**](cli-structure.md) - Detailed project structure and organization
- [**Implementation Roadmap**](roadmap.md) - Development phases and milestones

### 🏗️ **Architecture**
- [**Architecture Overview**](architecture/overview.md) - System design and component architecture
- [**Architecture Decisions**](architecture/decisions/register.md) - ADR register and decision history

#### **Foundation ADRs (ZEN-001)**
- [ADR-0001: Go Language Choice](architecture/decisions/ADR-0001-go-language-choice.md)
- [ADR-0002: Cobra CLI Framework](architecture/decisions/ADR-0002-cobra-cli-framework.md)
- [ADR-0003: Project Structure](architecture/decisions/ADR-0003-project-structure-organization.md)
- [ADR-0004: Configuration Management](architecture/decisions/ADR-0004-configuration-management-strategy.md)
- [ADR-0005: Structured Logging](architecture/decisions/ADR-0005-structured-logging-implementation.md)

### 📖 **User Documentation**

#### **CLI Commands**
- `zen --help` - Display help and available commands
- `zen version` - Show version and build information
- `zen init [directory]` - Initialize a new Zen workspace
- `zen config` - Display current configuration
- `zen status` - Show workspace status and health

#### **Configuration**
- **Configuration File**: `zen.yaml` (see [example](../configs/zen.example.yaml))
- **Environment Variables**: `ZEN_*` prefix (e.g., `ZEN_LOG_LEVEL=debug`)
- **Command-line Flags**: Global flags available for all commands

### 🛠️ **Development**

#### **Project Structure**
```
zen/
├── cmd/zen/                    # CLI Entry Point
├── internal/                   # Private Implementation
│   ├── cli/                   # Command Layer
│   ├── config/               # Configuration Management
│   ├── logging/              # Logging Infrastructure
│   └── ...                   # Future components
├── pkg/                       # Public APIs
│   ├── types/                # Common Types
│   └── errors/               # Error Handling
├── docs/                      # Documentation
├── .github/workflows/         # CI/CD Pipelines
└── configs/                   # Configuration Examples
```

#### **Build System**
- **Build**: `make build` - Build zen binary for current platform
- **Test**: `make test-unit` - Run unit tests with coverage
- **Cross-Platform**: `make build-all` - Build for all supported platforms
- **Coverage**: `make test-coverage` - Generate HTML coverage report
- **Lint**: `make lint` - Run linting checks
- **Security**: `make security` - Run security analysis

### 📋 **Stories & Planning**
- [**ZEN-001 Foundation**](stories/zen-001.md) - Go project setup and CLI framework

### 🎨 **Templates**
- [**Story Definition Template**](templates/story-definition.md) - Template for user stories

### 🔨 **Build & Deployment**
- [**Code Generation**](build/code-generation.md) - Automated code generation processes
- [**Feature Development**](build/code-feature.md) - Feature development workflow
- [**Story Definition**](build/define-story.md) - Story definition process

## 🚀 **Quick Start**

### **Installation**
```bash
# Clone the repository
git clone https://github.com/jonathandaddia/zen.git
cd zen

# Build the CLI
make build

# Initialize a workspace
./bin/zen init

# Check status
./bin/zen status
```

### **Development Setup**
```bash
# Install development tools
make dev-setup

# Run tests
make test-unit

# Run with debug logging
ZEN_DEBUG=true ./bin/zen --help
```

## 📊 **Current Status**

### ✅ **Completed (ZEN-001)**
- **Foundation**: Complete Go CLI framework with Cobra
- **Configuration**: Multi-source configuration with Viper
- **Logging**: Structured logging with multiple output formats
- **Build System**: Cross-platform builds and CI/CD pipeline
- **Testing**: Comprehensive unit tests with >80% coverage
- **Documentation**: Architecture decisions and user guides

### 🔄 **In Progress**
- Template system and content generation (ZEN-002)
- AI agent integration and orchestration
- External system integrations

### 📋 **Planned**
- Product management workflows
- Engineering workflow automation
- Quality gates and automation
- Plugin architecture and extensibility

## 🤝 **Contributing**

See [CONTRIBUTING.md](../CONTRIBUTING) for guidelines on contributing to the Zen CLI project.

## 📄 **License**

This project is licensed under the MIT License - see the [LICENSE](../LICENSE) file for details.

---

**Zen: Where AI intelligence meets product development productivity** 🧘‍♂️✨
