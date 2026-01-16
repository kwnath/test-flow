package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// MCP JSON-RPC types
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Workflow configuration (loaded from YAML)
type WorkflowConfig struct {
	Name        string       `yaml:"name" json:"name"`
	Description string       `yaml:"description" json:"description"`
	Steps       []StepConfig `yaml:"steps" json:"steps"`
}

type StepConfig struct {
	Name          string `yaml:"name" json:"name"`
	NeedsApproval bool   `yaml:"needs_approval" json:"needs_approval"`
	Instructions  string `yaml:"instructions" json:"instructions"`
}

// Workflow runtime state
type WorkflowState struct {
	ID                   string         `json:"id"`
	Task                 string         `json:"task"`
	CurrentStep          string         `json:"current_step"`
	Steps                []WorkflowStep `json:"steps"`
	WaitingForApproval   bool           `json:"waiting_for_approval"`
	VerificationCriteria []string       `json:"verification_criteria,omitempty"`
	CreatedAt            string         `json:"created_at"`
	UpdatedAt            string         `json:"updated_at"`
}

type WorkflowStep struct {
	Name          string `json:"name"`
	Status        string `json:"status"` // pending, in_progress, completed, blocked
	NeedsApproval bool   `json:"needs_approval"`
	Instructions  string `json:"instructions"`
}

type WorkflowEvent struct {
	Event      string `json:"event"`
	Type       string `json:"type"`
	WorkflowID string `json:"workflow_id"`
	Step       string `json:"step,omitempty"`
	NextStep   string `json:"next_step,omitempty"`
	Status     string `json:"status,omitempty"`
	Message    string `json:"message,omitempty"`
	Timestamp  string `json:"timestamp"`
}

var state *WorkflowState
var config *WorkflowConfig
var stateFile string
var configFile string

func main() {
	// Determine file locations
	cwd, _ := os.Getwd()
	stateFile = filepath.Join(cwd, "tmp", "workflow-state.json")
	configFile = filepath.Join(cwd, "workflow.yaml")

	// Load workflow configuration
	loadConfig()

	// Try to load existing state
	loadState()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			continue
		}

		resp := handleRequest(req)
		output, _ := json.Marshal(resp)
		fmt.Println(string(output))
	}
}

func loadConfig() {
	config = &WorkflowConfig{
		Name:        "default",
		Description: "Default workflow",
		Steps: []StepConfig{
			{Name: "plan", NeedsApproval: true, Instructions: "Explore the codebase and design your approach."},
			{Name: "execute", NeedsApproval: false, Instructions: "Implement the changes."},
			{Name: "verify", NeedsApproval: false, Instructions: "Run tests and verify criteria."},
			{Name: "pr", NeedsApproval: true, Instructions: "Create a pull request."},
			{Name: "complete", NeedsApproval: false, Instructions: "Summarize accomplishments."},
		},
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return // Use default config
	}

	var loaded WorkflowConfig
	if err := yaml.Unmarshal(data, &loaded); err == nil && len(loaded.Steps) > 0 {
		config = &loaded
	}
}

func handleRequest(req Request) Response {
	switch req.Method {
	case "initialize":
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]any{
					"tools": map[string]any{},
				},
				"serverInfo": map[string]any{
					"name":    "workflow-mcp",
					"version": "2.0.0",
				},
			},
		}

	case "tools/list":
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"tools": []map[string]any{
					{
						"name":        "workflow_init",
						"description": "Initialize a new workflow with a task description. Returns workflow config and first step instructions.",
						"inputSchema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"task": map[string]any{
									"type":        "string",
									"description": "Description of the task",
								},
							},
							"required": []string{"task"},
						},
					},
					{
						"name":        "workflow_status",
						"description": "Get current workflow status, progress, and step instructions",
						"inputSchema": map[string]any{
							"type":       "object",
							"properties": map[string]any{},
						},
					},
					{
						"name":        "workflow_step",
						"description": "Update workflow step status",
						"inputSchema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"step": map[string]any{
									"type":        "string",
									"description": "Step name",
								},
								"status": map[string]any{
									"type":        "string",
									"description": "New status (in_progress, completed, blocked)",
								},
							},
							"required": []string{"step", "status"},
						},
					},
					{
						"name":        "workflow_blocked",
						"description": "Mark workflow as blocked due to external dependencies (not for approval gates)",
						"inputSchema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"reason": map[string]any{
									"type":        "string",
									"description": "Reason for blocking (external dependency)",
								},
							},
							"required": []string{"reason"},
						},
					},
					{
						"name":        "workflow_next",
						"description": "Complete current step and move to next. Use when step work is done or approval is given.",
						"inputSchema": map[string]any{
							"type":       "object",
							"properties": map[string]any{},
						},
					},
					{
						"name":        "workflow_approve",
						"description": "Approve the current step (clears waiting_for_approval). Call this when user gives approval.",
						"inputSchema": map[string]any{
							"type":       "object",
							"properties": map[string]any{},
						},
					},
					{
						"name":        "workflow_set_criteria",
						"description": "Set verification criteria to be checked in the verify step",
						"inputSchema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"criteria": map[string]any{
									"type":        "array",
									"items":       map[string]any{"type": "string"},
									"description": "List of verification criteria (tests to run, checks to perform)",
								},
							},
							"required": []string{"criteria"},
						},
					},
				},
			},
		}

	case "tools/call":
		var params struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		json.Unmarshal(req.Params, &params)

		result := handleToolCall(params.Name, params.Arguments)
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"content": []map[string]any{
					{
						"type": "text",
						"text": result,
					},
				},
			},
		}

	default:
		return Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]any{},
		}
	}
}

func handleToolCall(name string, args map[string]any) string {
	switch name {
	case "workflow_init":
		return workflowInit(args["task"].(string))
	case "workflow_status":
		return workflowStatus()
	case "workflow_step":
		return workflowStep(args["step"].(string), args["status"].(string))
	case "workflow_blocked":
		return workflowBlocked(args["reason"].(string))
	case "workflow_next":
		return workflowNext()
	case "workflow_approve":
		return workflowApprove()
	case "workflow_set_criteria":
		criteria := []string{}
		if c, ok := args["criteria"].([]any); ok {
			for _, item := range c {
				if s, ok := item.(string); ok {
					criteria = append(criteria, s)
				}
			}
		}
		return workflowSetCriteria(criteria)
	default:
		return `{"error": "unknown tool"}`
	}
}

func workflowInit(task string) string {
	// Build steps from config
	steps := make([]WorkflowStep, len(config.Steps))
	for i, sc := range config.Steps {
		status := "pending"
		if i == 0 {
			status = "in_progress"
		}
		steps[i] = WorkflowStep{
			Name:          sc.Name,
			Status:        status,
			NeedsApproval: sc.NeedsApproval,
			Instructions:  sc.Instructions,
		}
	}

	firstStep := config.Steps[0]
	state = &WorkflowState{
		ID:                 fmt.Sprintf("wf_%d", time.Now().UnixNano()),
		Task:               task,
		CurrentStep:        firstStep.Name,
		Steps:              steps,
		WaitingForApproval: firstStep.NeedsApproval,
		CreatedAt:          time.Now().UTC().Format(time.RFC3339),
		UpdatedAt:          time.Now().UTC().Format(time.RFC3339),
	}
	saveState()

	event := WorkflowEvent{
		Event:      "workflow",
		Type:       "init",
		WorkflowID: state.ID,
		Step:       firstStep.Name,
		Status:     "in_progress",
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	output, _ := json.MarshalIndent(map[string]any{
		"workflow_id":          state.ID,
		"task":                 task,
		"current_step":         state.CurrentStep,
		"waiting_for_approval": state.WaitingForApproval,
		"instructions":         firstStep.Instructions,
		"steps":                state.Steps,
		"event":                event,
	}, "", "  ")
	return string(output)
}

func workflowStatus() string {
	if state == nil {
		return `{"error": "no workflow initialized", "hint": "call workflow_init first"}`
	}

	// Calculate progress
	completed := 0
	for _, s := range state.Steps {
		if s.Status == "completed" {
			completed++
		}
	}
	progress := float64(completed) / float64(len(state.Steps)) * 100

	// Get current step instructions
	var instructions string
	for _, s := range state.Steps {
		if s.Name == state.CurrentStep {
			instructions = s.Instructions
			break
		}
	}

	output, _ := json.MarshalIndent(map[string]any{
		"workflow_id":           state.ID,
		"task":                  state.Task,
		"current_step":          state.CurrentStep,
		"waiting_for_approval":  state.WaitingForApproval,
		"verification_criteria": state.VerificationCriteria,
		"progress":              fmt.Sprintf("%.0f%%", progress),
		"instructions":          instructions,
		"steps":                 state.Steps,
	}, "", "  ")
	return string(output)
}

func workflowStep(step, status string) string {
	if state == nil {
		return `{"error": "no workflow initialized"}`
	}

	// Update step status
	for i, s := range state.Steps {
		if s.Name == step {
			state.Steps[i].Status = status
			if status == "in_progress" {
				state.CurrentStep = step
				state.WaitingForApproval = state.Steps[i].NeedsApproval
			}
			break
		}
	}
	state.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	saveState()

	event := WorkflowEvent{
		Event:      "workflow",
		Type:       "step_update",
		WorkflowID: state.ID,
		Step:       step,
		Status:     status,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	output, _ := json.MarshalIndent(map[string]any{
		"updated":              true,
		"current_step":         state.CurrentStep,
		"waiting_for_approval": state.WaitingForApproval,
		"event":                event,
	}, "", "  ")
	return string(output)
}

func workflowBlocked(reason string) string {
	if state == nil {
		return `{"error": "no workflow initialized"}`
	}

	// Mark current step as blocked
	for i, s := range state.Steps {
		if s.Name == state.CurrentStep {
			state.Steps[i].Status = "blocked"
			break
		}
	}
	state.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	saveState()

	event := WorkflowEvent{
		Event:      "workflow",
		Type:       "blocked",
		WorkflowID: state.ID,
		Step:       state.CurrentStep,
		Status:     "blocked",
		Message:    reason,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	output, _ := json.MarshalIndent(map[string]any{
		"blocked":                  true,
		"step":                     state.CurrentStep,
		"reason":                   reason,
		"needs_human_intervention": true,
		"event":                    event,
	}, "", "  ")
	return string(output)
}

func workflowNext() string {
	if state == nil {
		return `{"error": "no workflow initialized"}`
	}

	// Find current step and move to next
	var previousStep string
	var nextStep string
	for i, s := range state.Steps {
		if s.Name == state.CurrentStep {
			previousStep = s.Name
			// Mark current as completed
			state.Steps[i].Status = "completed"
			// Find next step
			if i+1 < len(state.Steps) {
				nextStep = state.Steps[i+1].Name
				state.Steps[i+1].Status = "in_progress"
				state.CurrentStep = nextStep
				state.WaitingForApproval = state.Steps[i+1].NeedsApproval
			} else {
				state.CurrentStep = "done"
				state.WaitingForApproval = false
			}
			break
		}
	}
	state.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	saveState()

	// Get next step instructions
	var instructions string
	for _, s := range state.Steps {
		if s.Name == nextStep {
			instructions = s.Instructions
			break
		}
	}

	event := WorkflowEvent{
		Event:      "workflow",
		Type:       "step_complete",
		WorkflowID: state.ID,
		Step:       previousStep,
		NextStep:   nextStep,
		Status:     "in_progress",
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	output, _ := json.MarshalIndent(map[string]any{
		"previous_step":        previousStep,
		"current_step":         state.CurrentStep,
		"waiting_for_approval": state.WaitingForApproval,
		"instructions":         instructions,
		"event":                event,
	}, "", "  ")
	return string(output)
}

func workflowApprove() string {
	if state == nil {
		return `{"error": "no workflow initialized"}`
	}

	if !state.WaitingForApproval {
		return `{"error": "not waiting for approval", "hint": "current step does not require approval"}`
	}

	state.WaitingForApproval = false
	state.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	saveState()

	event := WorkflowEvent{
		Event:      "workflow",
		Type:       "approved",
		WorkflowID: state.ID,
		Step:       state.CurrentStep,
		Status:     "approved",
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	output, _ := json.MarshalIndent(map[string]any{
		"approved":             true,
		"step":                 state.CurrentStep,
		"waiting_for_approval": false,
		"hint":                 "Call workflow_next to proceed to the next step",
		"event":                event,
	}, "", "  ")
	return string(output)
}

func workflowSetCriteria(criteria []string) string {
	if state == nil {
		return `{"error": "no workflow initialized"}`
	}

	state.VerificationCriteria = criteria
	state.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	saveState()

	event := WorkflowEvent{
		Event:      "workflow",
		Type:       "criteria_set",
		WorkflowID: state.ID,
		Step:       state.CurrentStep,
		Message:    fmt.Sprintf("Set %d verification criteria", len(criteria)),
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	output, _ := json.MarshalIndent(map[string]any{
		"criteria_set": true,
		"criteria":     criteria,
		"event":        event,
	}, "", "  ")
	return string(output)
}

func saveState() {
	if state == nil {
		return
	}
	os.MkdirAll(filepath.Dir(stateFile), 0755)
	data, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile(stateFile, data, 0644)
}

func loadState() {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return
	}
	state = &WorkflowState{}
	json.Unmarshal(data, state)
}
