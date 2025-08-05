# Nomadic - Project Structure

This document outlines the proposed project structure for the Nomadic travel journal companion TUI application.

## Directory Structure

```
nomadic/
├── cmd/                    # Application entry points
│   └── nomadic/            # Main TUI application
│       └── main.go         # Entry point for the TUI
├── internal/               # Private application code
│   ├── config/             # Configuration management
│   │   └── config.go
│   ├── models/             # Core data models
│   │   ├── trip.go
│   │   ├── entry.go
│   │   └── expense.go
│   ├── storage/            # Data persistence
│   │   ├── json_store.go   # JSON file-based storage
│   │   └── store.go        # Storage interface
│   ├── llm/                # LLM integration
│   │   ├── client.go       # LLM API client
│   │   ├── prompts.go      # Prompt templates
│   │   └── processor.go    # Text processing utilities
│   └── ui/                 # TUI implementation with Bubbletea
│       ├── model.go        # Main Bubbletea model
│       ├── trip.go         # Trip-related views
│       ├── entry.go        # Entry-related views
│       └── expense.go      # Expense-related views
├── pkg/                    # Public libraries that can be used by external applications
│   ├── currency/           # Currency conversion utilities
│   └── formatter/          # Text formatting utilities
├── docs/                   # Documentation
│   └── usage.md
├── examples/               # Example configurations and usage
├── scripts/                # Build and utility scripts
├── test/                   # Integration and end-to-end tests
├── .gitignore
├── go.mod
├── go.sum
├── LICENSE
├── README.md
└── IMPLEMENTATION_PATH.md
```

## Key Files

### Main Application Entry Point

```go
// cmd/nomadic/main.go
package main

import (
	"fmt"
	"os"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/username/nomadic/internal/ui"
)

func main() {
	p := tea.NewProgram(ui.NewModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
```

### Core Data Models

```go
// internal/models/trip.go
package models

import (
	"time"
)

// Trip represents a travel journey with associated entries and expenses
type Trip struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Locations   []string   `json:"locations"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Entries     []Entry    `json:"entries"`
	Expenses    []Expense  `json:"expenses"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// NewTrip creates a new trip with the given title and locations
func NewTrip(title string, locations []string, startDate time.Time) *Trip {
	return &Trip{
		ID:        generateID(),
		Title:     title,
		Locations: locations,
		StartDate: startDate,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Entries:   []Entry{},
		Expenses:  []Expense{},
	}
}

// AddEntry adds a new entry to the trip
func (t *Trip) AddEntry(entry Entry) {
	t.Entries = append(t.Entries, entry)
	t.UpdatedAt = time.Now()
}

// AddExpense adds a new expense to the trip
func (t *Trip) AddExpense(expense Expense) {
	t.Expenses = append(t.Expenses, expense)
	t.UpdatedAt = time.Now()
}

// generateID creates a unique ID for the trip
func generateID() string {
	// Implementation details...
	return "trip-id"
}
```

```go
// internal/models/entry.go
package models

import (
	"time"
)

// Entry represents a journal entry in a trip
type Entry struct {
	ID        string            `json:"id"`
	TripID    string            `json:"trip_id"`
	Text      string            `json:"text"`
	Timestamp time.Time         `json:"timestamp"`
	Location  string            `json:"location,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"` // For LLM-extracted tags
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// NewEntry creates a new journal entry
func NewEntry(tripID, text string, timestamp time.Time) *Entry {
	return &Entry{
		ID:        generateEntryID(),
		TripID:    tripID,
		Text:      text,
		Timestamp: timestamp,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]string),
	}
}

// AddMetadata adds or updates metadata for the entry
func (e *Entry) AddMetadata(key, value string) {
	e.Metadata[key] = value
	e.UpdatedAt = time.Now()
}

// generateEntryID creates a unique ID for the entry
func generateEntryID() string {
	// Implementation details...
	return "entry-id"
}
```

```go
// internal/models/expense.go
package models

import (
	"time"
)

// Expense represents a financial expense during a trip
type Expense struct {
	ID          string    `json:"id"`
	TripID      string    `json:"trip_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewExpense creates a new expense record
func NewExpense(tripID string, amount float64, currency, category, description string, timestamp time.Time) *Expense {
	return &Expense{
		ID:          generateExpenseID(),
		TripID:      tripID,
		Amount:      amount,
		Currency:    currency,
		Category:    category,
		Description: description,
		Timestamp:   timestamp,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// generateExpenseID creates a unique ID for the expense
func generateExpenseID() string {
	// Implementation details...
	return "expense-id"
}
```

### Storage Interface

```go
// internal/storage/store.go
package storage

import (
	"github.com/username/nomadic/internal/models"
)

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

### TUI Models

```go
// internal/ui/model.go
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the main application model
type Model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

// NewModel creates a new application model
func NewModel() Model {
	return Model{
		choices: []string{
			"✈️  New Trip",
			"📔 View Journal",
			"💰 Expenses",
			"🛑 Quit",
		},
		selected: make(map[int]struct{}),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles user input and updates the model state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "w":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "s":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			selected := m.choices[m.cursor]
			if selected == "🛑 Quit" {
				return m, tea.Quit
			}
			// Handle other menu selections
		}
	}
	return m, nil
}

// View renders the UI
func (m Model) View() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Align(lipgloss.Center).
		Render("✈️  Nomadic – Your Travel Journal Companion")

	menu := ""
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = "👉"
		}
		menu += lipgloss.NewStyle().Render(cursor + " " + choice + "\n")
	}

	return title + "\n\n" + menu
}
```

```go
// internal/ui/trip.go
package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/username/nomadic/internal/models"
)

// TripModel represents the trip creation view
type TripModel struct {
	title     string
	locations []string
	startDate time.Time
	focused   int
	err       error
}

// NewTripModel creates a new trip creation model
func NewTripModel() TripModel {
	return TripModel{
		startDate: time.Now(),
		locations: []string{""},
	}
}

// Update handles user input for the trip creation view
func (m TripModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if m.title != "" && len(m.locations) > 0 && m.locations[0] != "" {
				// Create and save trip
				trip := models.NewTrip(m.title, m.locations, m.startDate)
				// Save trip implementation...
				
				// Return to main menu
				return NewModel(), nil
			}
		}
	}
	
	// Handle text input for fields based on focused field
	// Implementation details...
	
	return m, nil
}
```

## LLM Integration

```go
// internal/llm/client.go
package llm

import (
	"context"
	"errors"
)

// Provider represents the type of LLM provider
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderLocal     Provider = "local"
)

// Client defines the interface for LLM API interactions
type Client interface {
	Complete(ctx context.Context, prompt string, options CompletionOptions) (string, error)
}

// CompletionOptions contains parameters for the completion request
type CompletionOptions struct {
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
	Stop        []string `json:"stop"`
}

// NewClient creates a new LLM client based on the provider
func NewClient(provider Provider, apiKey string) (Client, error) {
	switch provider {
	case ProviderOpenAI:
		return newOpenAIClient(apiKey), nil
	case ProviderAnthropic:
		return newAnthropicClient(apiKey), nil
	case ProviderLocal:
		return newLocalClient(), nil
	default:
		return nil, errors.New("unsupported LLM provider")
	}
}
```

## Next Steps

1. Initialize the Go module and project structure
2. Implement the core data models
3. Create the storage interface and JSON implementation
4. Build the basic TUI with Bubbletea
5. Add tests for core functionality
6. Implement LLM integration for basic features