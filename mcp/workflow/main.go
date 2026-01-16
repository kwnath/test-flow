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

// Step metadata for approval and iteration
type StepMetadata struct {
	RequiresApproval bool   `json:"requires_approval"`
	AllowsIteration  bool   `json:"allows_iteration"`
	ApprovalPrompt   string `json:"approval_prompt,omitempty"`
}

// Artifact stores step outputs in a consistent structure
type Artifact struct {
	Type      string `json:"type"`                 // "plan", "criteria", "pr", "test_results", etc.
	Content   any    `json:"content"`              // flexible content (string, []string, map, etc.)
	Step      string `json:"step"`                 // which step created this
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// Workflow configuration (loaded from YAML)
type WorkflowConfig struct {
	Name        string       `yaml:"name" json:"name"`
	Description string       `yaml:"description" json:"description"`
	Steps       []StepConfig `yaml:"steps" json:"steps"`
}

type StepConfig struct {
	Name             string `yaml:"name" json:"name"`
	NeedsApproval    bool   `yaml:"needs_approval" json:"needs_approval"`
	AllowsIteration  bool   `yaml:"allows_iteration" json:"allows_iteration"`
	ApprovalPrompt   string `yaml:"approval_prompt" json:"approval_prompt,omitempty"`
	Instructions     string `yaml:"instructions" json:"instructions"`
}

// Workflow runtime state
type WorkflowState struct {
	ID                 string              `json:"id"`
	Task               string              `json:"task"`
	CurrentStep        string              `json:"current_step"`
	Steps              []WorkflowStep      `json:"steps"`
	WaitingForApproval bool                `json:"waiting_for_approval"`
	Artifacts          map[string]Artifact `json:"artifacts,omitempty"`
	IterationCount     int                 `json:"iteration_count"`
	IterationFeedback  []string            `json:"iteration_feedback,omitempty"`
	// PR tracking
	PRNumber         int    `json:"pr_number,omitempty"`
	LastCommentCheck string `json:"last_comment_check,omitempty"`
	LastCommentCount int    `json:"last_comment_count,omitempty"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type WorkflowStep struct {
	Name          string        `json:"name"`
	Status        string        `json:"status"` // pending, in_progress, awaiting_approval, completed, blocked
	NeedsApproval bool          `json:"needs_approval"`
	Instructions  string        `json:"instructions"`
	Metadata      *StepMetadata `json:"metadata,omitempty"`
}

type WorkflowEvent struct {
	Event          string `json:"event"`
	Type           string `json:"type"`
	WorkflowID     string `json:"workflow_id"`
	Step           string `json:"step,omitempty"`
	NextStep       string `json:"next_step,omitempty"`
	Status         string `json:"status,omitempty"`
	Message        string `json:"message,omitempty"`
	ApprovalPrompt string `json:"approval_prompt,omitempty"`
	CanIterate     bool   `json:"can_iterate,omitempty"`
	Timestamp      string `json:"timestamp"`
}

var state *WorkflowState
var config *WorkflowConfig
var stateFile string
var configFile string

// Default approval prompts for each step
var defaultApprovalPrompts = map[string]string{
	"plan":     "Review the implementation plan. Does this approach look correct? You can approve with /workflow-approve or request changes with /workflow-iterate <feedback>",
	"criteria": "Review the completion criteria. Are these the right things to verify? Approve with /workflow-approve or iterate with /workflow-iterate <feedback>",
	"review":   "PR review complete. Ready to merge? Approve with /workflow-approve to finish, or /workflow-iterate <feedback> for more changes.",
}

func main() {
	// Determine file locations
	cwd, _ := os.Getwd()
	homeDir, _ := os.UserHomeDir()
	stateFile = filepath.Join(homeDir, "state", "workflow_state.json")
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
			{Name: "plan", NeedsApproval: true, AllowsIteration: true, ApprovalPrompt: defaultApprovalPrompts["plan"], Instructions: "Explore the codebase and design your approach. Include diagrams to visualize architecture."},
			{Name: "criteria", NeedsApproval: true, AllowsIteration: true, ApprovalPrompt: defaultApprovalPrompts["criteria"], Instructions: "Define specific, measurable completion criteria."},
			{Name: "execute", NeedsApproval: false, AllowsIteration: true, Instructions: "Implement the changes."},
			{Name: "verify", NeedsApproval: false, AllowsIteration: true, Instructions: "Run tests and verify all criteria pass."},
			{Name: "pr", NeedsApproval: false, AllowsIteration: false, Instructions: "Create a pull request."},
			{Name: "review", NeedsApproval: true, AllowsIteration: true, ApprovalPrompt: defaultApprovalPrompts["review"], Instructions: "Monitor PR for comments, address feedback, check every 2 mins."},
			{Name: "complete", NeedsApproval: false, AllowsIteration: false, Instructions: "Summarize accomplishments."},
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
					"version": "2.1.0",
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
						"description": "Request to move to the next step. If step requires approval, sets status to awaiting_approval. Otherwise moves to next step.",
						"inputSchema": map[string]any{
							"type":       "object",
							"properties": map[string]any{},
						},
					},
					{
						"name":        "workflow_approve",
						"description": "Approve the current step and move to the next step. Only works when step is awaiting_approval.",
						"inputSchema": map[string]any{
							"type":       "object",
							"properties": map[string]any{},
						},
					},
					{
						"name":        "workflow_iterate",
						"description": "Provide feedback and iterate on the current step. Keeps you on the same step to revise based on feedback.",
						"inputSchema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"feedback": map[string]any{
									"type":        "string",
									"description": "Feedback for iteration - what needs to change",
								},
							},
							"required": []string{"feedback"},
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
					{
						"name":        "workflow_set_artifact",
						"description": "Store an artifact (plan, criteria, test results, etc.) in the workflow state. Artifacts are keyed by type and can be retrieved by vibe apps.",
						"inputSchema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"type": map[string]any{
									"type":        "string",
									"description": "Artifact type (e.g., 'plan', 'criteria', 'pr', 'test_results')",
								},
								"content": map[string]any{
									"description": "The artifact content (string, array, or object)",
								},
							},
							"required": []string{"type", "content"},
						},
					},
					{
						"name":        "workflow_set_pr",
						"description": "Set the PR number for tracking. Used by the review step to monitor comments.",
						"inputSchema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"pr_number": map[string]any{
									"type":        "integer",
									"description": "The pull request number",
								},
								"pr_url": map[string]any{
									"type":        "string",
									"description": "The pull request URL (optional)",
								},
							},
							"required": []string{"pr_number"},
						},
					},
					{
						"name":        "workflow_check_pr",
						"description": "Check if there are new PR comments since last check. Returns comment status and suggests next action.",
						"inputSchema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"comment_count": map[string]any{
									"type":        "integer",
									"description": "Current number of comments on the PR (from gh pr view)",
								},
							},
							"required": []string{"comment_count"},
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
	case "workflow_iterate":
		feedback := ""
		if f, ok := args["feedback"].(string); ok {
			feedback = f
		}
		return workflowIterate(feedback)
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
	case "workflow_set_plan":
		plan := ""
		if p, ok := args["plan"].(string); ok {
			plan = p
		}
		return workflowSetPlan(plan)
	case "workflow_set_artifact":
		artifactType := ""
		if t, ok := args["type"].(string); ok {
			artifactType = t
		}
		content := args["content"]
		return workflowSetArtifact(artifactType, content)
	case "workflow_set_pr":
		prNumber := 0
		if n, ok := args["pr_number"].(float64); ok {
			prNumber = int(n)
		}
		prURL := ""
		if u, ok := args["pr_url"].(string); ok {
			prURL = u
		}
		return workflowSetPR(prNumber, prURL)
	case "workflow_check_pr":
		commentCount := 0
		if c, ok := args["comment_count"].(float64); ok {
			commentCount = int(c)
		}
		return workflowCheckPR(commentCount)
	default:
		return `{"error": "unknown tool"}`
	}
}

func workflowInit(task string) string {
	// Build steps from config with metadata
	steps := make([]WorkflowStep, len(config.Steps))
	for i, sc := range config.Steps {
		status := "pending"
		if i == 0 {
			status = "in_progress"
		}

		// Build metadata
		metadata := &StepMetadata{
			RequiresApproval: sc.NeedsApproval,
			AllowsIteration:  sc.AllowsIteration,
			ApprovalPrompt:   sc.ApprovalPrompt,
		}

		// Use default approval prompt if not specified
		if metadata.RequiresApproval && metadata.ApprovalPrompt == "" {
			if prompt, ok := defaultApprovalPrompts[sc.Name]; ok {
				metadata.ApprovalPrompt = prompt
			}
		}

		steps[i] = WorkflowStep{
			Name:          sc.Name,
			Status:        status,
			NeedsApproval: sc.NeedsApproval,
			Instructions:  sc.Instructions,
			Metadata:      metadata,
		}
	}

	firstStep := config.Steps[0]
	state = &WorkflowState{
		ID:                 fmt.Sprintf("wf_%d", time.Now().UnixNano()),
		Task:               task,
		CurrentStep:        firstStep.Name,
		Steps:              steps,
		WaitingForApproval: false, // Not waiting yet - work must be done first
		Artifacts:          make(map[string]Artifact),
		IterationCount:     0,
		IterationFeedback:  []string{},
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
		CanIterate: firstStep.AllowsIteration,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	output, _ := json.MarshalIndent(map[string]any{
		"workflow_id":          state.ID,
		"task":                 task,
		"current_step":         state.CurrentStep,
		"waiting_for_approval": state.WaitingForApproval,
		"requires_approval":    firstStep.NeedsApproval,
		"allows_iteration":     firstStep.AllowsIteration,
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

	// Get current step info
	var instructions string
	var metadata *StepMetadata
	for _, s := range state.Steps {
		if s.Name == state.CurrentStep {
			instructions = s.Instructions
			metadata = s.Metadata
			break
		}
	}

	result := map[string]any{
		"workflow_id":          state.ID,
		"task":                 state.Task,
		"current_step":         state.CurrentStep,
		"waiting_for_approval": state.WaitingForApproval,
		"artifacts":            state.Artifacts,
		"iteration_count":      state.IterationCount,
		"iteration_feedback":   state.IterationFeedback,
		"progress":             fmt.Sprintf("%.0f%%", progress),
		"instructions":         instructions,
		"steps":                state.Steps,
	}

	// Add PR tracking if set
	if state.PRNumber > 0 {
		result["pr_number"] = state.PRNumber
		result["last_comment_check"] = state.LastCommentCheck
		result["last_comment_count"] = state.LastCommentCount
	}

	if metadata != nil {
		result["requires_approval"] = metadata.RequiresApproval
		result["allows_iteration"] = metadata.AllowsIteration
		if metadata.ApprovalPrompt != "" {
			result["approval_prompt"] = metadata.ApprovalPrompt
		}
	}

	output, _ := json.MarshalIndent(result, "", "  ")
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
				state.WaitingForApproval = false
				state.IterationCount = 0
				state.IterationFeedback = []string{}
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

	// Find current step
	var currentStepIdx int = -1
	var currentStep *WorkflowStep
	for i, s := range state.Steps {
		if s.Name == state.CurrentStep {
			currentStepIdx = i
			currentStep = &state.Steps[i]
			break
		}
	}

	if currentStep == nil {
		return `{"error": "current step not found"}`
	}

	// If step requires approval and is in_progress, set to awaiting_approval
	if currentStep.NeedsApproval && currentStep.Status == "in_progress" {
		state.Steps[currentStepIdx].Status = "awaiting_approval"
		state.WaitingForApproval = true
		state.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		saveState()

		approvalPrompt := ""
		canIterate := false
		if currentStep.Metadata != nil {
			approvalPrompt = currentStep.Metadata.ApprovalPrompt
			canIterate = currentStep.Metadata.AllowsIteration
		}

		event := WorkflowEvent{
			Event:          "workflow",
			Type:           "awaiting_approval",
			WorkflowID:     state.ID,
			Step:           currentStep.Name,
			Status:         "awaiting_approval",
			ApprovalPrompt: approvalPrompt,
			CanIterate:     canIterate,
			Timestamp:      time.Now().UTC().Format(time.RFC3339),
		}

		output, _ := json.MarshalIndent(map[string]any{
			"status":               "awaiting_approval",
			"step":                 currentStep.Name,
			"waiting_for_approval": true,
			"approval_prompt":      approvalPrompt,
			"can_iterate":          canIterate,
			"message":              "STOP AND WAIT for user approval. Do not proceed until user calls /workflow-approve or /workflow-iterate",
			"event":                event,
		}, "", "  ")
		return string(output)
	}

	// Step doesn't require approval or is already approved - move to next
	previousStep := currentStep.Name
	state.Steps[currentStepIdx].Status = "completed"

	var nextStep string
	var instructions string
	var requiresApproval bool
	var allowsIteration bool

	if currentStepIdx+1 < len(state.Steps) {
		nextStep = state.Steps[currentStepIdx+1].Name
		state.Steps[currentStepIdx+1].Status = "in_progress"
		state.CurrentStep = nextStep
		instructions = state.Steps[currentStepIdx+1].Instructions
		if state.Steps[currentStepIdx+1].Metadata != nil {
			requiresApproval = state.Steps[currentStepIdx+1].Metadata.RequiresApproval
			allowsIteration = state.Steps[currentStepIdx+1].Metadata.AllowsIteration
		}
	} else {
		state.CurrentStep = "done"
	}

	// Reset iteration tracking for new step
	state.WaitingForApproval = false
	state.IterationCount = 0
	state.IterationFeedback = []string{}
	state.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	saveState()

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
		"requires_approval":    requiresApproval,
		"allows_iteration":     allowsIteration,
		"instructions":         instructions,
		"event":                event,
	}, "", "  ")
	return string(output)
}

func workflowApprove() string {
	if state == nil {
		return `{"error": "no workflow initialized"}`
	}

	// Find current step
	var currentStepIdx int = -1
	var currentStep *WorkflowStep
	for i, s := range state.Steps {
		if s.Name == state.CurrentStep {
			currentStepIdx = i
			currentStep = &state.Steps[i]
			break
		}
	}

	if currentStep == nil {
		return `{"error": "current step not found"}`
	}

	if currentStep.Status != "awaiting_approval" {
		return `{"error": "step is not awaiting approval", "hint": "call workflow_next first to request approval", "current_status": "` + currentStep.Status + `"}`
	}

	// Mark current step as completed and move to next
	previousStep := currentStep.Name
	state.Steps[currentStepIdx].Status = "completed"

	var nextStep string
	var instructions string
	var requiresApproval bool
	var allowsIteration bool

	if currentStepIdx+1 < len(state.Steps) {
		nextStep = state.Steps[currentStepIdx+1].Name
		state.Steps[currentStepIdx+1].Status = "in_progress"
		state.CurrentStep = nextStep
		instructions = state.Steps[currentStepIdx+1].Instructions
		if state.Steps[currentStepIdx+1].Metadata != nil {
			requiresApproval = state.Steps[currentStepIdx+1].Metadata.RequiresApproval
			allowsIteration = state.Steps[currentStepIdx+1].Metadata.AllowsIteration
		}
	} else {
		state.CurrentStep = "done"
	}

	// Reset iteration tracking
	state.WaitingForApproval = false
	state.IterationCount = 0
	state.IterationFeedback = []string{}
	state.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	saveState()

	event := WorkflowEvent{
		Event:      "workflow",
		Type:       "approved",
		WorkflowID: state.ID,
		Step:       previousStep,
		NextStep:   nextStep,
		Status:     "approved",
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	output, _ := json.MarshalIndent(map[string]any{
		"approved":             true,
		"previous_step":        previousStep,
		"current_step":         state.CurrentStep,
		"waiting_for_approval": false,
		"requires_approval":    requiresApproval,
		"allows_iteration":     allowsIteration,
		"instructions":         instructions,
		"event":                event,
	}, "", "  ")
	return string(output)
}

func workflowIterate(feedback string) string {
	if state == nil {
		return `{"error": "no workflow initialized"}`
	}

	// Find current step
	var currentStepIdx int = -1
	var currentStep *WorkflowStep
	for i, s := range state.Steps {
		if s.Name == state.CurrentStep {
			currentStepIdx = i
			currentStep = &state.Steps[i]
			break
		}
	}

	if currentStep == nil {
		return `{"error": "current step not found"}`
	}

	// Check if iteration is allowed
	allowsIteration := false
	if currentStep.Metadata != nil {
		allowsIteration = currentStep.Metadata.AllowsIteration
	}

	if !allowsIteration {
		return `{"error": "iteration not allowed on this step", "step": "` + currentStep.Name + `"}`
	}

	// Increment iteration count and store feedback
	state.IterationCount++
	if feedback != "" {
		state.IterationFeedback = append(state.IterationFeedback, feedback)
	}

	// Set status back to in_progress
	state.Steps[currentStepIdx].Status = "in_progress"
	state.WaitingForApproval = false
	state.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	saveState()

	event := WorkflowEvent{
		Event:      "workflow",
		Type:       "iteration",
		WorkflowID: state.ID,
		Step:       currentStep.Name,
		Status:     "in_progress",
		Message:    feedback,
		CanIterate: true,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	output, _ := json.MarshalIndent(map[string]any{
		"iterated":           true,
		"step":               currentStep.Name,
		"iteration_count":    state.IterationCount,
		"feedback":           feedback,
		"all_feedback":       state.IterationFeedback,
		"instructions":       currentStep.Instructions,
		"message":            "Revise your work based on the feedback, then call workflow_next when ready for approval",
		"event":              event,
	}, "", "  ")
	return string(output)
}

func workflowSetCriteria(criteria []string) string {
	// Legacy function - now uses artifacts internally
	return workflowSetArtifact("criteria", criteria)
}

func workflowSetPlan(plan string) string {
	// Legacy function - now uses artifacts internally
	return workflowSetArtifact("plan", plan)
}

func workflowSetArtifact(artifactType string, content any) string {
	if state == nil {
		return `{"error": "no workflow initialized"}`
	}

	if state.Artifacts == nil {
		state.Artifacts = make(map[string]Artifact)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	artifact := Artifact{
		Type:      artifactType,
		Content:   content,
		Step:      state.CurrentStep,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// If artifact already exists, preserve CreatedAt
	if existing, ok := state.Artifacts[artifactType]; ok {
		artifact.CreatedAt = existing.CreatedAt
	}

	state.Artifacts[artifactType] = artifact
	state.UpdatedAt = now
	saveState()

	event := WorkflowEvent{
		Event:      "workflow",
		Type:       "artifact_set",
		WorkflowID: state.ID,
		Step:       state.CurrentStep,
		Message:    fmt.Sprintf("Artifact '%s' has been set", artifactType),
		Timestamp:  now,
	}

	output, _ := json.MarshalIndent(map[string]any{
		"artifact_set": true,
		"type":         artifactType,
		"step":         state.CurrentStep,
		"event":        event,
	}, "", "  ")
	return string(output)
}

func workflowSetPR(prNumber int, prURL string) string {
	if state == nil {
		return `{"error": "no workflow initialized"}`
	}

	state.PRNumber = prNumber
	state.LastCommentCheck = time.Now().UTC().Format(time.RFC3339)
	state.LastCommentCount = 0
	state.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	// Also store as artifact
	prArtifact := map[string]any{
		"number": prNumber,
		"url":    prURL,
	}
	if state.Artifacts == nil {
		state.Artifacts = make(map[string]Artifact)
	}
	state.Artifacts["pr"] = Artifact{
		Type:      "pr",
		Content:   prArtifact,
		Step:      state.CurrentStep,
		CreatedAt: state.UpdatedAt,
	}

	saveState()

	event := WorkflowEvent{
		Event:      "workflow",
		Type:       "pr_set",
		WorkflowID: state.ID,
		Step:       state.CurrentStep,
		Message:    fmt.Sprintf("PR #%d set for tracking", prNumber),
		Timestamp:  state.UpdatedAt,
	}

	output, _ := json.MarshalIndent(map[string]any{
		"pr_set":    true,
		"pr_number": prNumber,
		"pr_url":    prURL,
		"event":     event,
	}, "", "  ")
	return string(output)
}

func workflowCheckPR(commentCount int) string {
	if state == nil {
		return `{"error": "no workflow initialized"}`
	}

	if state.PRNumber == 0 {
		return `{"error": "no PR set", "hint": "call workflow_set_pr first"}`
	}

	now := time.Now().UTC()
	lastCheck, _ := time.Parse(time.RFC3339, state.LastCommentCheck)
	timeSinceLastCheck := now.Sub(lastCheck)

	hasNewComments := commentCount > state.LastCommentCount
	noNewCommentsTimeout := 1 * time.Minute

	// Update tracking
	state.LastCommentCheck = now.Format(time.RFC3339)
	previousCount := state.LastCommentCount
	state.LastCommentCount = commentCount
	state.UpdatedAt = now.Format(time.RFC3339)
	saveState()

	var action string
	var message string

	if hasNewComments {
		newCount := commentCount - previousCount
		action = "address_comments"
		message = fmt.Sprintf("Found %d new comment(s). Address the feedback, then check again.", newCount)
	} else if timeSinceLastCheck >= noNewCommentsTimeout {
		// No new comments for 1+ minute - ready to proceed
		action = "ready_for_approval"
		message = "No new comments for 1+ minute. Ready to call workflow_next for approval."
	} else {
		// Still within timeout window, keep waiting
		waitSeconds := int((noNewCommentsTimeout - timeSinceLastCheck).Seconds())
		action = "wait"
		message = fmt.Sprintf("No new comments yet. Wait %d seconds then check again.", waitSeconds)
	}

	event := WorkflowEvent{
		Event:      "workflow",
		Type:       "pr_check",
		WorkflowID: state.ID,
		Step:       state.CurrentStep,
		Message:    message,
		Timestamp:  now.Format(time.RFC3339),
	}

	output, _ := json.MarshalIndent(map[string]any{
		"pr_number":           state.PRNumber,
		"comment_count":       commentCount,
		"previous_count":      previousCount,
		"has_new_comments":    hasNewComments,
		"seconds_since_check": int(timeSinceLastCheck.Seconds()),
		"action":              action,
		"message":             message,
		"event":               event,
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
