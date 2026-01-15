package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
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

// Workflow types
type WorkflowState struct {
	ID          string         `json:"id"`
	Task        string         `json:"task"`
	CurrentStep string         `json:"current_step"`
	Steps       []WorkflowStep `json:"steps"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

type WorkflowStep struct {
	Name   string `json:"name"`
	Status string `json:"status"` // pending, in_progress, completed, blocked
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
var stateFile string

func main() {
	// Determine state file location
	cwd, _ := os.Getwd()
	stateFile = filepath.Join(cwd, "tmp", "workflow-state.json")

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
					"version": "1.0.0",
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
						"description": "Initialize a new workflow with a task description",
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
						"description": "Get current workflow status and progress",
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
									"description": "Step name (plan, criteria, execute, verify, pr, review)",
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
						"description": "Mark workflow as blocked, needing human intervention",
						"inputSchema": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"reason": map[string]any{
									"type":        "string",
									"description": "Reason for blocking",
								},
							},
							"required": []string{"reason"},
						},
					},
					{
						"name":        "workflow_next",
						"description": "Move to the next workflow step",
						"inputSchema": map[string]any{
							"type":       "object",
							"properties": map[string]any{},
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
	default:
		return `{"error": "unknown tool"}`
	}
}

func workflowInit(task string) string {
	state = &WorkflowState{
		ID:          fmt.Sprintf("wf_%d", time.Now().UnixNano()),
		Task:        task,
		CurrentStep: "plan",
		Steps: []WorkflowStep{
			{Name: "plan", Status: "in_progress"},
			{Name: "criteria", Status: "pending"},
			{Name: "execute", Status: "pending"},
			{Name: "verify", Status: "pending"},
			{Name: "pr", Status: "pending"},
			{Name: "review", Status: "pending"},
		},
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	saveState()

	event := WorkflowEvent{
		Event:      "workflow",
		Type:       "init",
		WorkflowID: state.ID,
		Step:       "plan",
		Status:     "in_progress",
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	output, _ := json.MarshalIndent(map[string]any{
		"workflow": state,
		"event":    event,
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

	output, _ := json.MarshalIndent(map[string]any{
		"workflow_id":  state.ID,
		"task":         state.Task,
		"current_step": state.CurrentStep,
		"progress":     fmt.Sprintf("%.0f%%", progress),
		"steps":        state.Steps,
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
		"updated":      true,
		"current_step": state.CurrentStep,
		"event":        event,
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
		"blocked":              true,
		"step":                 state.CurrentStep,
		"reason":               reason,
		"needs_human_intervention": true,
		"event":                event,
	}, "", "  ")
	return string(output)
}

func workflowNext() string {
	if state == nil {
		return `{"error": "no workflow initialized"}`
	}

	// Find current step and move to next
	var nextStep string
	for i, s := range state.Steps {
		if s.Name == state.CurrentStep {
			// Mark current as completed
			state.Steps[i].Status = "completed"
			// Find next pending step
			if i+1 < len(state.Steps) {
				nextStep = state.Steps[i+1].Name
				state.Steps[i+1].Status = "in_progress"
				state.CurrentStep = nextStep
			} else {
				state.CurrentStep = "done"
			}
			break
		}
	}
	state.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	saveState()

	event := WorkflowEvent{
		Event:      "workflow",
		Type:       "step_complete",
		WorkflowID: state.ID,
		Step:       state.CurrentStep,
		NextStep:   nextStep,
		Status:     "in_progress",
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	output, _ := json.MarshalIndent(map[string]any{
		"previous_step": state.CurrentStep,
		"current_step":  nextStep,
		"event":         event,
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
