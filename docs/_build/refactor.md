Refactor and implement the recommended structure changes:



```
pkg/clients/           # Public client interfaces
├── jira/              # Jira client (move from pkg/jira/)
├── git/              # Git client (move from pkg/git/)
├── ai/               # AI provider clients (OpenAI, Anthropic, etc.)
├── http/             # Shared HTTP utilities
└── types.go          # Common client types
```

```
internal/providers/    # Internal provider implementations
├── jira/             # Move from internal/integration/providers/
├── github/
├── openai/
├── anthropic/
└── azure/
```

