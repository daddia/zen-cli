# Draft Command

The `zen draft` command generates document templates populated with task manifest data from the zen-assets repository.

## Overview

The draft command integrates with the existing Zen CLI architecture to:

1. **Validate Context**: Ensures command is run within a task directory with a valid manifest.yaml
2. **Discover Templates**: Lists available activities from the zen-assets repository
3. **Process Templates**: Fetches Go templates and processes them with task-specific data
4. **Generate Output**: Creates formatted documents in the appropriate task directory

## Architecture

The command follows established Zen CLI patterns:

```
zen draft <activity>
    ↓
cmdutil.Factory (dependency injection)
    ↓
Asset Client (template fetching)
    ↓
Processor Factory (template processing)
    ↓
File System (output generation)
```

### Key Components

- **TaskManifest**: Represents task manifest.yaml structure
- **ProcessingContext**: Template processing context with task data
- **Processor Factory**: Handles different output formats (Markdown, YAML, JSON)
- **Helper Functions**: Manifest loading, path resolution, data mapping

## Usage

### Basic Usage

```bash
# Generate a feature specification
zen draft feature-spec

# Generate a user story
zen draft user-story

# Generate an epic definition
zen draft epic
```

### Advanced Options

```bash
# Preview template without generating file
zen draft feature-spec --preview

# Force overwrite existing file
zen draft roadmap --force

# Generate to custom path
zen draft api-docs --output ./docs/api.md
```

### Shell Completion

The command supports shell auto-completion for activity names:

```bash
zen draft <TAB>  # Shows available activities
zen draft feat<TAB>  # Completes to feature-spec
```

## Template Processing

### Data Mapping

The command maps task manifest data to template variables:

| Template Variable | Manifest Source | Description |
|------------------|-----------------|-------------|
| `TASK_ID` | `task.id` | Task identifier |
| `TASK_TITLE` | `task.title` | Task title |
| `TASK_TYPE` | `task.type` | Task type (story, bug, epic, etc.) |
| `OWNER_NAME` | `owner.name` | Task owner name |
| `OWNER_EMAIL` | `owner.email` | Task owner email |
| `TEAM_NAME` | `team.name` | Team name |
| `CREATED_DATE` | `dates.created` | Task creation date |
| `TARGET_DATE` | `dates.target` | Target completion date |
| `CURRENT_STAGE` | `workflow.current_stage` | Current workflow stage |
| `BUSINESS_CRITERIA` | `success_criteria.business` | Business success criteria |
| `TECHNICAL_CRITERIA` | `success_criteria.technical` | Technical success criteria |

### Template Formats

The command supports multiple output formats:

- **Markdown** (`.md`): Documentation, specifications, reports
- **YAML** (`.yaml`): Configuration files, data structures
- **JSON** (`.json`): API responses, structured data

### Processing Pipeline

1. **Template Parsing**: Go template syntax with variable substitution
2. **Data Population**: Task manifest data mapped to template variables
3. **Format Processing**: Format-specific validation and formatting
4. **Output Generation**: Formatted content written to file

## Error Handling

The command provides comprehensive error handling:

### Error Types

- **Invalid Input**: Unknown activity, malformed arguments
- **Configuration**: Asset client initialization failures
- **Network**: Template fetching failures
- **Permission**: File system access issues
- **Processing**: Template rendering failures

### Error Messages

- Clear, actionable error messages
- Suggestions for similar activities when activity not found
- Detailed context for debugging issues

### Example Error Output

```sh
Error: Unknown activity 'feature-specification'. Did you mean: feature-spec?

Error: File feature-spec.md already exists. Use --force to overwrite or --preview to see content

Error: Failed to load task manifest: no task manifest.yaml found in current directory or parent directories
```

## Testing

### Test Coverage

The command includes comprehensive test coverage:

- **Unit Tests**: Individual function testing
- **Integration Tests**: Full command workflow testing
- **Mock Testing**: Asset client and processor mocking
- **Error Testing**: All error conditions and edge cases

### Test Structure

```
pkg/cmd/draft/
├── draft.go          # Main command implementation
├── helpers.go        # Helper functions
└── draft_test.go     # Comprehensive test suite
```

### Running Tests

```bash
# Run all draft command tests
go test ./pkg/cmd/draft/

# Run tests with coverage
go test -cover ./pkg/cmd/draft/

# Run integration tests
go test -v ./pkg/cmd/draft/ -run TestRunDraftIntegration
```

## Performance

### Benchmarks

- **Template Generation**: < 500ms (P95), < 200ms (P50)
- **Memory Usage**: < 50MB peak during processing
- **Cache Performance**: < 50ms for cached templates

### Optimization

- **Template Caching**: Reduces repeated fetching overhead
- **Lazy Loading**: Templates loaded only when needed
- **Processor Reuse**: Factory pattern minimizes initialization

## Security

### Input Validation

- Activity name validation against known activities
- File path sanitization to prevent directory traversal
- Template content validation before processing

### Data Handling

- No sensitive data logged (API keys, tokens, passwords)
- Secure file permissions (0644 for generated files)
- Proper error handling without information leakage

## Integration

### Existing Systems

The draft command integrates with:

- **Asset Client**: Template fetching and caching
- **Workspace Manager**: Task context detection
- **Template Engine**: Go template processing
- **CLI Framework**: Cobra command structure

### Extension Points

- **Custom Processors**: Add support for new output formats
- **Template Functions**: Extend Go template function library
- **Data Sources**: Additional data sources beyond task manifest
- **Output Formats**: New template formats and processors

## Future Enhancements

### Planned Features

- **Variable Override**: Command-line variable substitution
- **Batch Generation**: Multiple template generation
- **Custom Templates**: User-defined template support
- **Template Validation**: Schema-based template validation

### Architecture Improvements

- **Plugin System**: Extensible processor architecture
- **Template Registry**: Local template management
- **Performance Monitoring**: Detailed performance metrics
- **Caching Strategy**: Advanced caching with TTL and invalidation

## Contributing

### Development Guidelines

1. Follow existing code patterns and conventions
2. Maintain comprehensive test coverage (>90%)
3. Include proper error handling and validation
4. Update documentation for new features
5. Follow Zen CLI design guidelines for user experience

### Code Quality

- Use `gofmt` for code formatting
- Follow Go best practices and idioms
- Include comprehensive error handling
- Write clear, self-documenting code
- Maintain backward compatibility
