# Integration Guide for Vibe Coding Apps

This guide explains how external applications can integrate with the workflow system to display progress, plans, and handle approvals.

## Overview

The workflow MCP provides structured task management with:
- **State persistence** in `tmp/workflow-state.json`
- **Structured events** emitted during workflow transitions
- **Approval gates** that pause for human review

## Reading Workflow State

Poll or watch `tmp/workflow-state.json` for the current state:

```json
{
  "id": "wf_1737045123456789",
  "task": "Add user authentication",
  "current_step": "plan",
  "waiting_for_approval": true,
  "implementation_plan": "## Plan\n\n### Architecture\n...",
  "verification_criteria": ["Tests pass", "No errors"],
  "iteration_count": 1,
  "iteration_feedback": ["Add error handling section"],
  "steps": [
    {
      "name": "plan",
      "status": "awaiting_approval",
      "needs_approval": true,
      "metadata": {
        "requires_approval": true,
        "allows_iteration": true,
        "approval_prompt": "Review the plan..."
      }
    },
    {
      "name": "criteria",
      "status": "pending",
      "needs_approval": true
    }
  ]
}
```

## Key Fields to Display

| Field | Description | When to Show |
|-------|-------------|--------------|
| `task` | What the user asked for | Always |
| `current_step` | Active step name | Always |
| `waiting_for_approval` | Whether human input needed | Always |
| `implementation_plan` | The design/plan (markdown) | When set |
| `verification_criteria` | List of things to verify | When set |
| `iteration_count` | How many revisions | During approval |
| `iteration_feedback` | All feedback given | During approval |
| `steps[].status` | Step progress | Progress bar |

## Step Statuses

| Status | Meaning | UI Suggestion |
|--------|---------|---------------|
| `pending` | Not started | Gray/dimmed |
| `in_progress` | Currently working | Blue/animated |
| `awaiting_approval` | Needs human review | Yellow/attention |
| `completed` | Done | Green/checkmark |
| `blocked` | External blocker | Red/warning |

## Listening for Events

Events are emitted to stdout as JSON. Parse Claude's output for lines containing:

```json
{"event": "workflow", "type": "...", ...}
```

### Event Types

**`init`** - Workflow started
```json
{
  "event": "workflow",
  "type": "init",
  "workflow_id": "wf_123",
  "step": "plan",
  "status": "in_progress"
}
```

**`awaiting_approval`** - Needs human review (SHOW APPROVAL UI)
```json
{
  "event": "workflow",
  "type": "awaiting_approval",
  "step": "plan",
  "approval_prompt": "Review the implementation plan...",
  "can_iterate": true
}
```

**`approved`** - Human approved, moving forward
```json
{
  "event": "workflow",
  "type": "approved",
  "step": "plan",
  "next_step": "criteria"
}
```

**`iteration`** - Human requested changes
```json
{
  "event": "workflow",
  "type": "iteration",
  "step": "plan",
  "message": "Add error handling section"
}
```

**`plan_set`** - Plan was stored (UPDATE PLAN DISPLAY)
```json
{
  "event": "workflow",
  "type": "plan_set",
  "message": "Implementation plan has been set"
}
```

**`criteria_set`** - Criteria were stored
```json
{
  "event": "workflow",
  "type": "criteria_set",
  "message": "Set 3 verification criteria"
}
```

**`step_complete`** - Step finished, moving to next
```json
{
  "event": "workflow",
  "type": "step_complete",
  "step": "execute",
  "next_step": "verify"
}
```

## Handling Approvals

When `waiting_for_approval: true`, show approval UI:

### Option 1: Buttons
```
┌─────────────────────────────────────────┐
│ 📋 Plan Ready for Review                │
├─────────────────────────────────────────┤
│                                         │
│ [Show Plan]                             │
│                                         │
│ ┌─────────┐  ┌──────────────────────┐  │
│ │ Approve │  │ Request Changes...   │  │
│ └─────────┘  └──────────────────────┘  │
└─────────────────────────────────────────┘
```

### Option 2: Chat Commands
Tell the user they can type:
- `/workflow-approve` - Approve and continue
- `/workflow-iterate <feedback>` - Request changes

### Triggering Approval
Send to Claude:
```
/workflow-approve
```

### Triggering Iteration
Send to Claude:
```
/workflow-iterate Please add error handling to the plan
```

## Displaying the Plan

The `implementation_plan` field contains markdown. Render it with:
- Code blocks (may contain ASCII diagrams)
- Mermaid diagrams (```mermaid blocks)
- Headers, lists, etc.

Example plan content:
```markdown
## Implementation Plan

### Current Architecture
```
┌─────────┐     ┌─────────┐
│ Frontend│────▶│   API   │
└─────────┘     └─────────┘
```

### Proposed Changes
1. Add auth middleware
2. Create login endpoint
3. Add JWT validation

### Files to Modify
- src/middleware/auth.ts
- src/routes/login.ts
```

## Progress Visualization

Build a progress bar from `steps` array:

```
Plan ✓ → Criteria ✓ → Execute ● → Verify ○ → PR ○ → Complete ○
[====================|          ] 40%
```

Calculate: `completed_steps / total_steps * 100`

## Recommended UI Flow

1. **On workflow start**: Show task and step progress
2. **On `plan_set`**: Display the plan in a panel/modal
3. **On `awaiting_approval`**: Show approval buttons prominently
4. **On `approved`**: Animate transition to next step
5. **On `iteration`**: Show feedback was received, plan being revised
6. **On `criteria_set`**: Show verification checklist
7. **On `step_complete`**: Update progress bar

## Example: React Component State

```typescript
interface WorkflowDisplay {
  task: string;
  currentStep: string;
  waitingForApproval: boolean;
  plan: string | null;
  criteria: string[];
  progress: number;
  steps: Array<{
    name: string;
    status: 'pending' | 'in_progress' | 'awaiting_approval' | 'completed' | 'blocked';
  }>;
}

// Poll state file or parse events to update
function parseWorkflowState(json: string): WorkflowDisplay {
  const state = JSON.parse(json);
  const completed = state.steps.filter(s => s.status === 'completed').length;
  return {
    task: state.task,
    currentStep: state.current_step,
    waitingForApproval: state.waiting_for_approval,
    plan: state.implementation_plan || null,
    criteria: state.verification_criteria || [],
    progress: (completed / state.steps.length) * 100,
    steps: state.steps,
  };
}
```

## File Locations

| File | Purpose |
|------|---------|
| `tmp/workflow-state.json` | Current workflow state |
| `workflow.yaml` | Workflow configuration |

## Tips

1. **Poll frequency**: Check state file every 1-2 seconds during active work
2. **Event parsing**: Events appear inline in Claude's output, parse JSON objects with `"event": "workflow"`
3. **Plan display**: Use a collapsible panel - plans can be long
4. **Approval prominence**: Make approval buttons very visible when `waiting_for_approval: true`
5. **Iteration history**: Show `iteration_count` and `iteration_feedback` so users see revision history
