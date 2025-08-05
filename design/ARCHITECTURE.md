# Nomadic - Architecture and Integration Approach

This document outlines the architecture and integration approach for the Nomadic travel journal companion application, with a focus on how LLM capabilities are integrated into the system.

## System Architecture

### High-Level Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│                 │     │                 │     │                 │
│  TUI Interface  │────▶│  Core Services  │────▶│   Data Store    │
│                 │     │                 │     │                 │
└─────────────────┘     └────────┬────────┘     └─────────────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │                 │
                        │   LLM Service   │
                        │                 │
                        └─────────────────┘
```

### Component Breakdown

1. **TUI Interface**
   - User interaction layer
   - Interactive terminal UI using Bubbletea
   - Styled output using Lipgloss

2. **Core Services**
   - Trip management
   - Journal entry management
   - Expense tracking
   - Business logic

3. **LLM Service**
   - Prompt management
   - LLM API integration
   - Response processing
   - Caching

4. **Data Store**
   - Local file-based storage (JSON)
   - Future: Database storage

## Detailed Component Design

### TUI Interface

The TUI interface is built using the Bubbletea library, providing an interactive terminal user interface with styled components:

```
┌─────────────────────────────────────────────┐
│                                             │
│  ✈️  Nomadic – Your Travel Journal Companion │
│                                             │
│  👉 ✈️  New Trip                             │
│     📔 View Journal                         │
│     💰 Expenses                             │
│     🛑 Quit                                 │
│                                             │
└─────────────────────────────────────────────┘
```

The interface is composed of Bubbletea models that implement the tea.Model interface with Init, Update, and View methods:

```go
type model struct {
    choices  []string
    cursor   int
    selected map[int]struct{}
}

func (m model) Init() tea.Cmd {
    // Initialize the model
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle user input and update state
}

func (m model) View() string {
    // Render the UI
}
```

Text styling is handled using the Lipgloss library, which provides a declarative API for styling terminal output.

### Core Services

The core services implement the business logic of the application:

1. **Trip Service**
   - Create, read, update, delete trips
   - Manage trip metadata
   - Link entries and expenses to trips

2. **Entry Service**
   - Create, read, update, delete journal entries
   - Manage entry metadata
   - Support for rich text and media (future)

3. **Expense Service**
   - Create, read, update, delete expenses
   - Categorization and currency handling
   - Aggregation and reporting

4. **Analysis Service**
   - Integration point with LLM service
   - Orchestrates data preparation for LLM
   - Processes and formats LLM responses

### LLM Service

The LLM service is designed with the following principles:

1. **Abstraction**: The service provides a clean interface that abstracts away the details of specific LLM providers.

2. **Configurability**: Support for different LLM providers (OpenAI, Anthropic, local models) with configurable parameters.

3. **Prompt Management**: Structured approach to managing and versioning prompts.

4. **Caching**: Efficient caching to reduce API costs and improve performance.

5. **Fallbacks**: Graceful degradation when LLM services are unavailable.

#### LLM Service Interface

```go
// LLMService defines the interface for LLM-powered features
type LLMService interface {
    // SummarizeEntries generates a summary of journal entries
    SummarizeEntries(ctx context.Context, entries []models.Entry, options SummaryOptions) (string, error)
    
    // GenerateReflectivePrompts creates reflective questions based on entries
    GenerateReflectivePrompts(ctx context.Context, entries []models.Entry, options ReflectionOptions) ([]string, error)
    
    // AnalyzeEntryMetadata extracts metadata from an entry
    AnalyzeEntryMetadata(ctx context.Context, entry models.Entry) (map[string]string, error)
    
    // RecommendPlaces suggests places based on past entries
    RecommendPlaces(ctx context.Context, entries []models.Entry, currentLocation string) ([]Recommendation, error)
    
    // AnalyzeExpenses provides insights on expense patterns
    AnalyzeExpenses(ctx context.Context, expenses []models.Expense, options ExpenseAnalysisOptions) (ExpenseInsights, error)
}
```

#### Prompt Management

Prompts are managed as templates with variable substitution:

```go
// PromptTemplate represents a template for an LLM prompt
type PromptTemplate struct {
    Name        string
    Description string
    Template    string
    Version     string
}

// PromptManager handles the loading and rendering of prompt templates
type PromptManager interface {
    // GetPrompt retrieves a prompt template by name
    GetPrompt(name string) (PromptTemplate, error)
    
    // RenderPrompt renders a prompt template with the given data
    RenderPrompt(template PromptTemplate, data interface{}) (string, error)
}
```

Prompt templates are stored as embedded resources in the application, allowing for versioning and updates.

#### LLM Client

The LLM client provides a unified interface to different LLM providers:

```go
// LLMClient defines the interface for LLM API interactions
type LLMClient interface {
    // Complete generates a completion for the given prompt
    Complete(ctx context.Context, prompt string, options CompletionOptions) (string, error)
}

// CompletionOptions contains parameters for the completion request
type CompletionOptions struct {
    MaxTokens   int
    Temperature float64
    TopP        float64
    Stop        []string
}
```

Implementations are provided for different LLM providers:

- OpenAIClient: Integration with OpenAI API
- AnthropicClient: Integration with Anthropic API
- LocalClient: Integration with locally hosted models

### Data Store

The data store is responsible for persisting application data:

```go
// Store defines the interface for data persistence
type Store interface {
    // Trip operations
    SaveTrip(trip *models.Trip) error
    GetTrip(id string) (*models.Trip, error)
    ListTrips() ([]*models.Trip, error)
    DeleteTrip(id string) error
    
    // Entry operations
    SaveEntry(entry *models.Entry) error
    GetEntry(id string) (*models.Entry, error)
    ListEntriesByTrip(tripID string) ([]*models.Entry, error)
    
    // Expense operations
    SaveExpense(expense *models.Expense) error
    GetExpense(id string) (*models.Expense, error)
    ListExpensesByTrip(tripID string) ([]*models.Expense, error)
}
```

Initial implementation uses JSON files for storage, with a future path to database storage.

## Integration Patterns

### LLM Integration Flow

The typical flow for LLM-powered features follows this pattern:

1. **Data Collection**: Gather relevant data from the data store
2. **Data Preparation**: Format and filter data for the LLM
3. **Prompt Generation**: Render the appropriate prompt template with data
4. **LLM Invocation**: Send the prompt to the LLM service
5. **Response Processing**: Parse and validate the LLM response
6. **Result Presentation**: Format and present the results to the user

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│             │     │             │     │             │     │             │
│  Collect    │────▶│  Prepare    │────▶│  Generate   │────▶│  Invoke     │
│  Data       │     │  Data       │     │  Prompt     │     │  LLM        │
│             │     │             │     │             │     │             │
└─────────────┘     └─────────────┘     └─────────────┘     └──────┬──────┘
                                                                   │
┌─────────────┐     ┌─────────────┐                               │
│             │     │             │                               │
│  Present    │◀────│  Process    │◀──────────────────────────────┘
│  Results    │     │  Response   │
│             │     │             │
└─────────────┘     └─────────────┘
```

### Error Handling and Fallbacks

LLM integration includes robust error handling:

1. **API Failures**: Handle network errors, rate limits, and service outages
2. **Response Validation**: Ensure responses meet expected format and quality
3. **Fallback Mechanisms**: Provide degraded but functional experience when LLM is unavailable
4. **User Feedback**: Clear communication about errors and fallbacks

```go
func (s *llmService) SummarizeEntries(ctx context.Context, entries []models.Entry, options SummaryOptions) (string, error) {
    // Attempt to use LLM for summary
    summary, err := s.generateLLMSummary(ctx, entries, options)
    if err == nil {
        return summary, nil
    }
    
    // Log the error
    s.logger.Warn("Failed to generate LLM summary, using fallback", "error", err)
    
    // Fallback to basic summary
    return s.generateBasicSummary(entries), nil
}

func (s *llmService) generateBasicSummary(entries []models.Entry) string {
    // Simple fallback that doesn't require LLM
    if len(entries) == 0 {
        return "No entries available."
    }
    
    // Extract date range
    startDate := entries[0].Timestamp
    endDate := entries[0].Timestamp
    for _, entry := range entries {
        if entry.Timestamp.Before(startDate) {
            startDate = entry.Timestamp
        }
        if entry.Timestamp.After(endDate) {
            endDate = entry.Timestamp
        }
    }
    
    // Generate basic summary
    return fmt.Sprintf(
        "Journal contains %d entries from %s to %s.",
        len(entries),
        startDate.Format("2006-01-02"),
        endDate.Format("2006-01-02"),
    )
}
```

### Caching Strategy

To optimize performance and reduce API costs, a multi-level caching strategy is implemented:

1. **Request Caching**: Cache identical LLM requests to avoid duplicate API calls
2. **Result Caching**: Store processed results for common operations
3. **Invalidation**: Clear caches when underlying data changes

```go
type CachingLLMClient struct {
    underlying LLMClient
    cache      Cache
}

func (c *CachingLLMClient) Complete(ctx context.Context, prompt string, options CompletionOptions) (string, error) {
    // Generate cache key
    cacheKey := generateCacheKey(prompt, options)
    
    // Check cache
    if cachedResult, found := c.cache.Get(cacheKey); found {
        return cachedResult.(string), nil
    }
    
    // Call underlying client
    result, err := c.underlying.Complete(ctx, prompt, options)
    if err != nil {
        return "", err
    }
    
    // Cache result
    c.cache.Set(cacheKey, result, cache.DefaultExpiration)
    
    return result, nil
}
```

## Configuration Management

The application uses a layered configuration approach:

1. **Default Configuration**: Embedded defaults for all settings
2. **Configuration File**: User-specific configuration in `~/.nomadic/config.yaml`
3. **Environment Variables**: Override settings with environment variables
4. **Command-line Flags**: Override settings with command-line flags

LLM-specific configuration includes:

```yaml
llm:
  provider: openai  # openai, anthropic, local
  model: gpt-4      # Model identifier
  api_key: ""       # API key (can be set via NOMADIC_LLM_API_KEY env var)
  temperature: 0.7  # Creativity level (0.0-1.0)
  max_tokens: 500   # Maximum response length
  cache:
    enabled: true
    ttl: 24h        # Cache time-to-live
```

## Security Considerations

### API Key Management

API keys for LLM services are handled securely:

1. **Storage**: Keys are stored in the user's configuration file with appropriate permissions
2. **Environment Variables**: Keys can be provided via environment variables
3. **Keyring Integration**: Future support for system keyring/credential store

### Data Privacy

User data privacy is a priority:

1. **Local-First**: Data is stored locally by default
2. **Minimal Data Sharing**: Only necessary data is sent to LLM APIs
3. **Data Sanitization**: Personal information is redacted before sending to LLMs
4. **Transparency**: Clear documentation about what data is sent to external services

## Performance Optimization

### Token Optimization

LLM costs are directly tied to token usage, so optimization is important:

1. **Prompt Engineering**: Efficient prompts that minimize token usage
2. **Data Filtering**: Send only relevant data to the LLM
3. **Chunking**: Break large datasets into manageable chunks
4. **Response Length Control**: Set appropriate max token limits

### Response Time Optimization

To ensure a responsive user experience:

1. **Asynchronous Processing**: Long-running LLM operations run asynchronously
2. **Progress Indication**: Clear feedback during processing
3. **Caching**: Reuse results when possible
4. **Timeout Handling**: Graceful handling of slow responses

## Extensibility

The architecture is designed for extensibility:

1. **Plugin System**: Future support for user-created plugins
2. **Custom Prompts**: Allow users to customize or create prompts
3. **Provider Abstraction**: Easy addition of new LLM providers
4. **Feature Flags**: Gradual rollout of new features

## Deployment and Distribution

### TUI Application Distribution

The TUI application is distributed as:

1. **Binary Releases**: Pre-built binaries for major platforms
2. **Package Managers**: Installation via Homebrew, apt, etc.
3. **Go Install**: Direct installation via `go install`

### Configuration

Initial setup process:

1. **Interactive Setup**: Guided setup for first-time users
2. **API Key Configuration**: Secure storage of LLM API keys
3. **Default Settings**: Sensible defaults with customization options

## Future Architecture Evolution

### Web and Mobile Interfaces

Future expansion beyond TUI:

1. **Web Interface**: Browser-based UI with the same core functionality
2. **Mobile Apps**: Native mobile experience with offline capabilities
3. **Shared Backend**: Common core services across all platforms

### Advanced LLM Integration

Evolution of LLM capabilities:

1. **Fine-tuned Models**: Custom models trained on travel journaling
2. **Multi-modal Support**: Integration with image and audio analysis
3. **Conversational Interface**: Natural language interaction with journal data

### Cloud Synchronization

Future cloud capabilities:

1. **Secure Sync**: End-to-end encrypted synchronization
2. **Collaboration**: Shared trips and journals
3. **Backup and Restore**: Secure cloud backup

## Conclusion

The Nomadic architecture provides a solid foundation for a travel journal companion with integrated LLM capabilities. The design emphasizes:

1. **User-Centric**: Focus on user needs and experience
2. **Privacy-First**: Respect for user data and privacy
3. **Reliability**: Robust error handling and fallbacks
4. **Extensibility**: Clear paths for future growth
5. **Performance**: Efficient use of resources

This architecture enables the seamless integration of LLM capabilities while maintaining a responsive, reliable application that respects user privacy and provides valuable insights from travel journal data.