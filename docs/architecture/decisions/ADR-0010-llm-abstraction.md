---
status: Accepted
date: 2025-09-16
decision-makers: Development Team, Architecture Team
consulted: AI/ML Team, DevOps Team
informed: Product Team, Engineering Leadership
---

# ADR-0010 - LLM Provider Abstraction

## Context and Problem Statement

The Zen CLI platform requires a flexible abstraction layer for integrating multiple LLM providers (OpenAI, Anthropic, Azure OpenAI, local models) to enable vendor flexibility, cost optimization, and specialized model selection. The abstraction must normalize differences in API interfaces, token counting, streaming responses, and error handling while maintaining high performance and enabling provider-specific optimizations. This layer is critical for avoiding vendor lock-in and enabling intelligent model routing based on task requirements and cost constraints.

## Decision Drivers

* **Vendor Flexibility**: Ability to switch providers or use multiple providers simultaneously without application code changes
* **Cost Optimization**: Dynamic provider and model selection based on cost, latency, and capability requirements  
* **API Normalization**: Consistent interface despite varying provider APIs, authentication methods, and response formats
* **Performance Requirements**: Minimal abstraction overhead, streaming response support, connection pooling and reuse
* **Reliability**: Robust error handling, rate limiting, timeout management, and automatic retry logic
* **Feature Parity**: Support for all required LLM features (chat completion, embeddings, function calling) across providers

## Considered Options

1. **Strategy Pattern with Unified Interface**
2. **Adapter Pattern with Provider-Specific Implementations** 
3. **Proxy Pattern with Runtime Provider Selection**
4. **Direct Integration with Provider SDKs**

## Decision Outcome

Chosen option: "Strategy Pattern with Unified Interface", because it provides the cleanest separation between provider implementations and consumer code while enabling runtime provider selection and configuration. The strategy pattern naturally supports different provider capabilities while maintaining a consistent programming interface.

### Consequences

**Good:**
- Clean separation of concerns enables independent provider implementation and testing
- Runtime provider selection enables dynamic optimization based on workload and cost requirements
- Consistent interface simplifies agent implementation and reduces cognitive overhead for developers
- Easy to add new providers or modify existing implementations without affecting consumer code

**Bad:**
- Abstraction layer adds performance overhead for API calls and response processing
- Complex to implement feature compatibility across providers with different capabilities  
- Risk of lowest-common-denominator interface that limits access to provider-specific features

### Confirmation

Performance benchmarks showing <5% overhead compared to direct provider SDK usage, integration tests validating all provider implementations against unified interface contract, and cost analysis demonstrating effective provider selection optimization.

## Pros and Cons of the Options

### Strategy Pattern with Unified Interface

A common interface that all LLM providers implement, allowing runtime selection of providers while maintaining consistent API calls and response handling.

**Good:**
- Clear separation between provider implementations and business logic
- Runtime configurability for provider selection and optimization
- Easy to test each provider implementation independently
- Natural extensibility for new providers and capabilities

**Neutral:**
- Moderate implementation complexity requiring careful interface design

**Bad:**
- Performance overhead from abstraction layer and interface indirection
- Challenging to support provider-specific advanced features uniformly

### Adapter Pattern with Provider-Specific Implementations

Individual adapter classes for each provider that translate between the provider's native API and a common internal interface, preserving provider-specific features.

**Good:**
- Direct mapping to provider APIs minimizes impedance mismatch
- Can expose provider-specific features and optimizations where needed
- Clear encapsulation of provider-specific authentication and error handling

**Bad:**
- Less consistent programming interface across different providers
- More complex client code that must handle provider differences
- Difficult to implement dynamic provider selection and fallback logic

### Proxy Pattern with Runtime Provider Selection

A proxy layer that dynamically routes requests to appropriate providers based on configuration, load balancing, or request characteristics without exposing provider differences.

**Good:**
- Transparent provider switching without client code awareness
- Excellent support for load balancing, failover, and A/B testing scenarios
- Can implement sophisticated routing logic based on request characteristics

**Bad:**
- Complex implementation requiring deep understanding of all provider APIs
- Difficult to expose provider-specific features and capabilities
- Risk of abstraction leaks when providers behave differently

### Direct Integration with Provider SDKs

Using each provider's official SDK directly in application code without abstraction layers, handling provider differences explicitly in business logic.

**Good:**
- Maximum performance with no abstraction overhead
- Full access to provider-specific features and optimizations
- Minimal implementation complexity for simple use cases

**Bad:**
- Tight coupling to specific provider APIs creates vendor lock-in
- Complex client code must handle different APIs, authentication, and error patterns
- No support for dynamic provider selection or cost optimization strategies

## More Information

- Related ADRs: [ADR-0009](ADR-0009-agent-orchestration.md), [ADR-0011](ADR-0011-workflow-management.md)
- Implementation Location: `internal/agents/providers/`
- Provider Documentation: OpenAI API, Anthropic Claude API, Azure OpenAI Service
- Follow-ups: Cost tracking framework, streaming response optimization
