package ai

import (
	"fmt"
	"sync"
	"time"
)

// CostTracker tracks cumulative token usage and costs across multiple API calls
type CostTracker struct {
	mu sync.RWMutex

	// Cumulative totals
	totalInputTokens  int
	totalOutputTokens int
	totalCost         float64

	// Per-phase tracking
	phases map[string]*PhaseStats

	// Current phase
	currentPhase string

	// Model for cost calculation
	model Model

	// Start time
	startTime time.Time

	// Callback for cost updates
	onUpdate func(update CostUpdate)
}

// PhaseStats tracks stats for a specific migration phase
type PhaseStats struct {
	Name         string
	InputTokens  int
	OutputTokens int
	Cost         float64
	Calls        int
	StartTime    time.Time
	EndTime      time.Time
}

// CostUpdate represents a cost update notification
type CostUpdate struct {
	Phase        string
	InputTokens  int
	OutputTokens int
	PhaseCost    float64
	TotalCost    float64
	TotalInput   int
	TotalOutput  int
}

// NewCostTracker creates a new cost tracker
func NewCostTracker(model Model) *CostTracker {
	return &CostTracker{
		model:     model,
		phases:    make(map[string]*PhaseStats),
		startTime: time.Now(),
	}
}

// SetUpdateCallback sets a callback function for cost updates
func (t *CostTracker) SetUpdateCallback(callback func(CostUpdate)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onUpdate = callback
}

// StartPhase begins tracking a new phase
func (t *CostTracker) StartPhase(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.currentPhase = name
	t.phases[name] = &PhaseStats{
		Name:      name,
		StartTime: time.Now(),
	}
}

// EndPhase ends the current phase
func (t *CostTracker) EndPhase() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.currentPhase != "" {
		if phase, ok := t.phases[t.currentPhase]; ok {
			phase.EndTime = time.Now()
		}
	}
	t.currentPhase = ""
}

// RecordUsage records token usage from an API call
func (t *CostTracker) RecordUsage(inputTokens, outputTokens int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Calculate cost
	inputCost := float64(inputTokens) * t.model.InputCostPer1M / 1_000_000
	outputCost := float64(outputTokens) * t.model.OutputCostPer1M / 1_000_000
	callCost := inputCost + outputCost

	// Update totals
	t.totalInputTokens += inputTokens
	t.totalOutputTokens += outputTokens
	t.totalCost += callCost

	// Update phase stats
	if t.currentPhase != "" {
		if phase, ok := t.phases[t.currentPhase]; ok {
			phase.InputTokens += inputTokens
			phase.OutputTokens += outputTokens
			phase.Cost += callCost
			phase.Calls++
		}
	}

	// Notify callback
	if t.onUpdate != nil {
		var phaseCost float64
		if phase, ok := t.phases[t.currentPhase]; ok {
			phaseCost = phase.Cost
		}
		t.onUpdate(CostUpdate{
			Phase:        t.currentPhase,
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
			PhaseCost:    phaseCost,
			TotalCost:    t.totalCost,
			TotalInput:   t.totalInputTokens,
			TotalOutput:  t.totalOutputTokens,
		})
	}
}

// RecordFromResponse records usage from an AnalysisResponse
func (t *CostTracker) RecordFromResponse(resp *AnalysisResponse) {
	if resp != nil {
		t.RecordUsage(resp.InputTokens, resp.OutputTokens)
	}
}

// GetTotals returns the cumulative totals
func (t *CostTracker) GetTotals() (inputTokens, outputTokens int, cost float64) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.totalInputTokens, t.totalOutputTokens, t.totalCost
}

// GetPhaseStats returns stats for a specific phase
func (t *CostTracker) GetPhaseStats(name string) (*PhaseStats, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	stats, ok := t.phases[name]
	if !ok {
		return nil, false
	}
	// Return a copy
	copy := *stats
	return &copy, true
}

// GetAllPhaseStats returns stats for all phases
func (t *CostTracker) GetAllPhaseStats() map[string]*PhaseStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make(map[string]*PhaseStats, len(t.phases))
	for k, v := range t.phases {
		copy := *v
		result[k] = &copy
	}
	return result
}

// GetSummary returns a formatted summary of all costs
func (t *CostTracker) GetSummary() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	duration := time.Since(t.startTime)

	summary := fmt.Sprintf("\n╔══════════════════════════════════════════════════════════╗\n")
	summary += fmt.Sprintf("║                    Cost Summary                          ║\n")
	summary += fmt.Sprintf("╠══════════════════════════════════════════════════════════╣\n")
	summary += fmt.Sprintf("║  Model: %-48s ║\n", t.model.Name)
	summary += fmt.Sprintf("║  Duration: %-45s ║\n", duration.Round(time.Second))
	summary += fmt.Sprintf("╠══════════════════════════════════════════════════════════╣\n")

	// Phase breakdown
	for _, phase := range t.phases {
		if phase.Calls > 0 {
			summary += fmt.Sprintf("║  %-20s %8d in / %8d out  $%.4f ║\n",
				phase.Name+":", phase.InputTokens, phase.OutputTokens, phase.Cost)
		}
	}

	summary += fmt.Sprintf("╠══════════════════════════════════════════════════════════╣\n")
	summary += fmt.Sprintf("║  TOTAL: %12d input + %8d output = $%.4f    ║\n",
		t.totalInputTokens, t.totalOutputTokens, t.totalCost)
	summary += fmt.Sprintf("╚══════════════════════════════════════════════════════════╝\n")

	return summary
}

// FormatCost formats a cost value for display
func FormatCost(cost float64) string {
	if cost < 0.01 {
		return fmt.Sprintf("$%.4f", cost)
	}
	return fmt.Sprintf("$%.2f", cost)
}

// FormatTokens formats token counts for display
func FormatTokens(tokens int) string {
	if tokens >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(tokens)/1_000_000)
	}
	if tokens >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(tokens)/1_000)
	}
	return fmt.Sprintf("%d", tokens)
}

// EstimateMigrationCost provides a rough cost estimate for a migration
func EstimateMigrationCost(model Model, fileCount int, avgFileSizeBytes int) (minCost, maxCost float64) {
	// Assumptions:
	// - Each file is processed once for analysis
	// - Output is roughly 30-50% of input for code transformations
	// - Context overhead adds ~20% to input

	avgTokensPerFile := avgFileSizeBytes / 4 // ~4 bytes per token
	totalInputTokens := fileCount * avgTokensPerFile * 12 / 10 // +20% context

	// Min estimate: efficient processing, minimal output
	minOutputTokens := totalInputTokens * 3 / 10
	minCost = float64(totalInputTokens)*model.InputCostPer1M/1_000_000 +
		float64(minOutputTokens)*model.OutputCostPer1M/1_000_000

	// Max estimate: verbose output, retries
	maxOutputTokens := totalInputTokens * 5 / 10
	maxInputWithRetries := totalInputTokens * 15 / 10 // +50% for retries
	maxCost = float64(maxInputWithRetries)*model.InputCostPer1M/1_000_000 +
		float64(maxOutputTokens)*model.OutputCostPer1M/1_000_000

	return minCost, maxCost
}
