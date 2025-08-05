# Nomadic - Testing Strategy for LLM-Powered Features

This document outlines the testing strategy for the LLM-powered features in the Nomadic travel journal companion application.

## Testing Challenges

LLM-powered features present unique testing challenges:

1. **Non-deterministic outputs**: LLMs can produce different outputs for the same input, making traditional assertion-based testing difficult.
2. **Quality assessment**: Evaluating the quality of generated text requires subjective judgment.
3. **API dependencies**: Testing against real LLM APIs can be slow and costly.
4. **Prompt evolution**: As prompts evolve, tests need to be updated.
5. **Edge cases**: LLMs may handle edge cases unpredictably.

## Testing Approach

### 1. Unit Testing

#### Prompt Template Testing
- **Test prompt template rendering**: Ensure variables are correctly substituted.
- **Test prompt length calculation**: Verify token counting for budget management.

```go
func TestDailySummaryPromptRendering(t *testing.T) {
    data := PromptData{
        Date:     "2023-05-15",
        Location: "Kyoto, Japan",
        Entries:  "Sample entry text",
    }
    
    tmpl, err := template.New("dailySummary").Parse(DailySummaryPrompt)
    require.NoError(t, err)
    
    var promptBuf bytes.Buffer
    err = tmpl.Execute(&promptBuf, data)
    require.NoError(t, err)
    
    prompt := promptBuf.String()
    assert.Contains(t, prompt, "Date: 2023-05-15")
    assert.Contains(t, prompt, "Location: Kyoto, Japan")
    assert.Contains(t, prompt, "Sample entry text")
}
```

#### LLM Client Mocking
- Create mock implementations of the LLM client interface for testing.
- Use recorded responses for consistent testing.

```go
type MockLLMClient struct {
    responses map[string]string
}

func NewMockLLMClient() *MockLLMClient {
    return &MockLLMClient{
        responses: map[string]string{
            "daily_summary_kyoto": "The traveler arrived in Kyoto and stayed at a traditional ryokan near Gion district. They visited Kinkaku-ji (Golden Pavilion) in the afternoon, admiring the golden reflection on the pond despite the crowds. The day concluded with dinner at a local izakaya where they enjoyed yakitori and sake while meeting fellow travelers who shared tips about Nara.",
            // Add more canned responses for different test scenarios
        },
    }
}

func (m *MockLLMClient) Complete(ctx context.Context, prompt string, options CompletionOptions) (string, error) {
    // Simple matching based on keywords in the prompt
    if strings.Contains(prompt, "Kyoto") {
        return m.responses["daily_summary_kyoto"], nil
    }
    
    return "Default mock response", nil
}
```

### 2. Integration Testing

#### API Integration Tests
- Test the integration with actual LLM APIs using a small set of representative inputs.
- Use API keys for test environments with lower costs.
- Cache responses to reduce API calls during repeated test runs.

```go
func TestLiveAPIIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Only run if API key is available
    apiKey := os.Getenv("OPENAI_TEST_API_KEY")
    if apiKey == "" {
        t.Skip("OPENAI_TEST_API_KEY not set, skipping integration test")
    }
    
    client := NewOpenAIClient(apiKey)
    
    // Test with a simple, consistent prompt
    result, err := client.Complete(context.Background(), "Summarize: The weather is sunny.", CompletionOptions{
        MaxTokens: 50,
        Temperature: 0.0, // Use 0 temperature for more deterministic results in tests
    })
    
    require.NoError(t, err)
    assert.NotEmpty(t, result)
    // Basic validation that the response is a summary
    assert.True(t, len(result) < len("Summarize: The weather is sunny."))
}
```

#### Response Caching for Tests
- Implement a caching layer for test runs to avoid repeated API calls.

```go
type CachingLLMClient struct {
    underlying Client
    cacheDir   string
}

func (c *CachingLLMClient) Complete(ctx context.Context, prompt string, options CompletionOptions) (string, error) {
    cacheKey := fmt.Sprintf("%x", md5.Sum([]byte(prompt)))
    cachePath := filepath.Join(c.cacheDir, cacheKey)
    
    // Check cache
    if data, err := ioutil.ReadFile(cachePath); err == nil {
        return string(data), nil
    }
    
    // Call API
    result, err := c.underlying.Complete(ctx, prompt, options)
    if err != nil {
        return "", err
    }
    
    // Cache result
    if err := ioutil.WriteFile(cachePath, []byte(result), 0644); err != nil {
        log.Printf("Warning: Failed to cache LLM response: %v", err)
    }
    
    return result, nil
}
```

### 3. Snapshot Testing

- Compare LLM outputs against stored "golden" responses.
- Allow for controlled updates of golden files when prompt templates change.

```go
func TestSummarizeEntriesSnapshot(t *testing.T) {
    entries := []Entry{
        // Test entries...
    }
    
    // Use deterministic client for snapshot tests
    client := NewDeterministicTestClient()
    
    summary, err := SummarizeEntriesWithClient(client, entries, testDate, "Kyoto, Japan")
    require.NoError(t, err)
    
    // Compare with golden file
    goldenPath := filepath.Join("testdata", "golden", "kyoto_summary.txt")
    
    if os.Getenv("UPDATE_GOLDEN") == "1" {
        err = ioutil.WriteFile(goldenPath, []byte(summary), 0644)
        require.NoError(t, err)
        return
    }
    
    expected, err := ioutil.ReadFile(goldenPath)
    require.NoError(t, err)
    
    assert.Equal(t, string(expected), summary)
}
```

### 4. Semantic Testing

- Implement tests that check for semantic properties rather than exact matches.
- Use embeddings or simpler heuristics to compare meaning rather than exact text.

```go
func TestSummarySemanticProperties(t *testing.T) {
    entries := []Entry{
        // Test entries about a visit to Paris...
    }
    
    summary, err := SummarizeEntries(entries, testDate, "Paris, France")
    require.NoError(t, err)
    
    // Check that key entities are mentioned
    assert.True(t, containsAny(summary, []string{"Paris", "France", "French"}))
    
    // Check that the summary is shorter than the combined entries
    totalEntryLength := 0
    for _, entry := range entries {
        totalEntryLength += len(entry.Text)
    }
    assert.True(t, len(summary) < totalEntryLength)
    
    // Check reading level is appropriate
    readability := calculateReadabilityScore(summary)
    assert.True(t, readability > 40 && readability < 70) // Flesch Reading Ease
}

func containsAny(text string, keywords []string) bool {
    text = strings.ToLower(text)
    for _, keyword := range keywords {
        if strings.Contains(text, strings.ToLower(keyword)) {
            return true
        }
    }
    return false
}
```

### 5. Property-Based Testing

- Define properties that should hold for all outputs.
- Generate random inputs and verify the properties.

```go
func TestSummaryProperties(t *testing.T) {
    // Define property: Summary should never be longer than the combined entries
    property := func(entries []Entry) bool {
        if len(entries) == 0 {
            return true
        }
        
        // Ensure entries have the same date
        date := entries[0].Timestamp
        for i := range entries {
            entries[i].Timestamp = date
        }
        
        summary, err := SummarizeEntries(entries, date, "Test Location")
        if err != nil {
            return false
        }
        
        totalLength := 0
        for _, entry := range entries {
            totalLength += len(entry.Text)
        }
        
        return len(summary) < totalLength
    }
    
    // Run the property test
    config := quick.Config{MaxCount: 50}
    if err := quick.Check(property, &config); err != nil {
        t.Error("Property failed:", err)
    }
}
```

### 6. User Evaluation Tests

- Implement a framework for human evaluation of LLM outputs.
- Collect and track user feedback on generated content.

```go
type UserEvaluationResult struct {
    PromptID      string
    InputData     map[string]string
    Output        string
    Rating        int // 1-5
    Feedback      string
    EvaluatorID   string
    EvaluationDate time.Time
}

func RecordUserEvaluation(result UserEvaluationResult) error {
    // Store evaluation results in database
    // Track metrics over time
    // ...
}
```

## Test Data Management

### 1. Synthetic Test Data

- Create a diverse set of synthetic journal entries for testing.
- Include edge cases like very short/long entries, entries with unusual content, etc.

```go
func GenerateSyntheticEntries() []Entry {
    return []Entry{
        // Normal entry
        {
            ID:        "synthetic-1",
            TripID:    "synthetic-trip",
            Text:      "Visited the museum today. The exhibits were fascinating, especially the ancient artifacts section.",
            Timestamp: time.Date(2023, 6, 10, 14, 30, 0, 0, time.UTC),
            Location:  "Rome, Italy",
        },
        // Very short entry
        {
            ID:        "synthetic-2",
            TripID:    "synthetic-trip",
            Text:      "Tired today. Stayed in hotel.",
            Timestamp: time.Date(2023, 6, 10, 20, 15, 0, 0, time.UTC),
            Location:  "Rome, Italy",
        },
        // Very long entry
        {
            ID:        "synthetic-3",
            TripID:    "synthetic-trip",
            Text:      "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum." + 
                      "Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt.",
            Timestamp: time.Date(2023, 6, 10, 22, 0, 0, 0, time.UTC),
            Location:  "Rome, Italy",
        },
        // Entry with special characters and emojis
        {
            ID:        "synthetic-4",
            TripID:    "synthetic-trip",
            Text:      "Amazing day! 😍 The café had the best espresso ☕ I've ever tasted. Can't wait to come back! #blessed #travel",
            Timestamp: time.Date(2023, 6, 10, 16, 45, 0, 0, time.UTC),
            Location:  "Rome, Italy",
        },
    }
}
```

### 2. Anonymized Real Data

- Create test datasets from anonymized real user data (when available).
- Ensure personal information is removed or replaced.

## Continuous Integration

### 1. Test Automation

- Automate unit and integration tests in CI pipeline.
- Skip API-dependent tests in regular CI runs, but run them nightly.

```yaml
# Example GitHub Actions workflow
name: Nomadic Tests

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
  schedule:
    - cron: '0 0 * * *'  # Run nightly

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - name: Set up Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.18
    
    - name: Unit tests
      run: go test -v ./... -short
    
    - name: Integration tests
      if: github.event_name == 'schedule'  # Only run on nightly schedule
      env:
        OPENAI_TEST_API_KEY: ${{ secrets.OPENAI_TEST_API_KEY }}
      run: go test -v ./... -tags=integration
```

### 2. Quality Metrics

- Track metrics for LLM-generated content:
  - Response time
  - Token usage
  - User satisfaction ratings
  - Error rates

## Debugging Tools

### 1. Prompt Logging

- Log prompts and responses for debugging.
- Implement a debug mode that shows the full prompt sent to the LLM.

```go
type LoggingLLMClient struct {
    underlying Client
    logger     *log.Logger
}

func (c *LoggingLLMClient) Complete(ctx context.Context, prompt string, options CompletionOptions) (string, error) {
    c.logger.Printf("LLM Request:\nPrompt: %s\nOptions: %+v\n", prompt, options)
    
    start := time.Now()
    result, err := c.underlying.Complete(ctx, prompt, options)
    duration := time.Since(start)
    
    if err != nil {
        c.logger.Printf("LLM Error: %v (took %v)\n", err, duration)
        return "", err
    }
    
    c.logger.Printf("LLM Response (took %v):\n%s\n", duration, result)
    return result, nil
}
```

### 2. Response Analysis

- Implement tools to analyze LLM responses for quality and consistency.
- Track token usage and response times.

## Conclusion

Testing LLM-powered features requires a multi-faceted approach that combines traditional testing methods with specialized techniques for handling non-deterministic outputs. By implementing this testing strategy, we can ensure that Nomadic's LLM features are reliable, high-quality, and provide value to users.

The testing approach should evolve as the application matures and as LLM technology advances. Regular reviews of test effectiveness and user feedback will help refine both the testing strategy and the LLM implementation.