# Integration Guide for Vibe Coding Apps

This guide explains how external applications can integrate with the workflow system to display progress, plans, and handle approvals.

## Overview

The workflow MCP provides structured task management with:
- **State persistence** in `~/state/workflow_state.json`
- **Structured events** emitted during workflow transitions
- **Approval gates** that pause for human review

## Reading Workflow State

Poll or watch `~/state/workflow_state.json` for the current state:

```json
{
  "id": "wf_1737045123456789",
  "task": "Add user authentication",
  "current_step": "plan",
  "waiting_for_approval": true,
  "artifacts": {
    "plan": {
      "type": "plan",
      "content": "## Plan\n\n### Architecture\n...",
      "step": "plan",
      "created_at": "2025-01-15T12:00:00Z"
    },
    "criteria": {
      "type": "criteria",
      "content": [
        "- [x] All CRUD operations work",
        "- [x] Error handling for invalid inputs",
        "- [ ] Documentation complete"
      ],
      "step": "criteria",
      "created_at": "2025-01-15T12:05:00Z"
    },
    "pr": {
      "type": "pr",
      "content": {"number": 123, "url": "https://github.com/..."},
      "step": "pr",
      "created_at": "2025-01-15T12:30:00Z"
    }
  },
  "iteration_count": 1,
  "iteration_feedback": ["Add error handling section"],
  "pr_number": 123,
  "last_comment_check": "2025-01-15T12:35:00Z",
  "last_comment_count": 2,
  "steps": [
    {
      "name": "plan",
      "status": "completed",
      "needs_approval": true,
      "metadata": {
        "requires_approval": true,
        "allows_iteration": true,
        "approval_prompt": "Review the plan..."
      }
    },
    {
      "name": "review",
      "status": "in_progress",
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
| `artifacts.plan.content` | The design/plan (markdown) | When set |
| `artifacts.criteria.content` | Checklist with `- [ ]` / `- [x]` | When set |
| `artifacts.pr.content` | PR number and URL | During PR/review |
| `iteration_count` | How many revisions | During approval |
| `iteration_feedback` | All feedback given | During approval |
| `pr_number` | PR being tracked | During review step |
| `steps[].status` | Step progress | Progress bar |

## Artifacts Model

Artifacts provide a **consistent structure** for storing any workflow output:

```typescript
interface Artifact {
  type: string;      // "plan", "criteria", "pr", "test_results", etc.
  content: any;      // string, array, or object
  step: string;      // which step created it
  created_at: string;
  updated_at?: string;
}
```

Access artifacts by type:
- `artifacts.clarifications.content` - Questions asked and answers received
- `artifacts.plan.content` - The implementation plan (string, markdown)
- `artifacts.criteria.content` - Verification checklist (array of `- [ ]` / `- [x]` strings)
- `artifacts.pr.content` - PR info (object with `number`, `url`, `branch`)
- `artifacts.summary.content` - Goal progress summary (see below)

### Clarifications

When Claude asks clarifying questions, they're stored for display:

```json
{
  "artifacts": {
    "clarifications": {
      "content": {
        "questions": [
          {"id": "app_type", "question": "What type of todo app?", "answer": "CLI"},
          {"id": "language", "question": "Which language?", "answer": "Node.js"},
          {"id": "storage", "question": "How to store data?", "answer": "JSON file"}
        ]
      }
    }
  }
}
```

Display example:
```
┌─────────────────────────────────────────────────┐
│ Clarifications                                  │
├─────────────────────────────────────────────────┤
│ App Type: CLI                                   │
│ Language: Node.js                               │
│ Storage: JSON file                              │
└─────────────────────────────────────────────────┘
```

Artifacts are extensible - new types can be added without code changes.

### Rendering Criteria Checklist

Criteria items use markdown checkbox format. Render as interactive checkboxes:

```typescript
// Parse criteria items
const criteria = artifacts.criteria?.content || [];
// ["- [x] All CRUD operations work", "- [ ] Docs complete"]

// Render each item
criteria.map(item => {
  const checked = item.includes('[x]');
  const text = item.replace(/^- \[.\] /, '');
  return { checked, text };
});
// [{checked: true, text: "All CRUD operations work"}, ...]
```

Display example:
```
☑ All CRUD operations work
☑ Error handling for invalid inputs
☐ Documentation complete
```

## Goal Summary

The `summary` artifact provides **context-rich status** for anyone picking up the project:

```json
{
  "artifacts": {
    "summary": {
      "type": "summary",
      "content": "**Goal:** Add user authentication\n\n**Context:** JWT-based auth with refresh tokens, storing sessions in Redis\n\n**Done:**\n- Designed auth flow with middleware\n- Implemented login/logout endpoints\n- Added token refresh logic\n\n**Now:** Verifying implementation (step 4/7)\n\n**Next:** Create PR",
      "step": "verify",
      "updated_at": "2025-01-16T12:30:00Z"
    }
  }
}
```

### Summary Fields

| Field | Description |
|-------|-------------|
| **Goal** | Original task |
| **Context** | What it is (tech, approach) |
| **Done** | Brief list of completed work |
| **Now** | Current step and position |
| **Next** | What comes after |

### Display Example

Render `content` as markdown:

```
┌─────────────────────────────────────────────────┐
│ Goal: Add user authentication                   │
│                                                 │
│ Context: JWT-based auth with refresh tokens,    │
│ storing sessions in Redis                       │
│                                                 │
│ Done:                                           │
│ • Designed auth flow with middleware            │
│ • Implemented login/logout endpoints            │
│ • Added token refresh logic                     │
│                                                 │
│ Now: Verifying implementation (step 4/7)        │
│ Next: Create PR                                 │
└─────────────────────────────────────────────────┘
```

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

**`artifact_set`** - Any artifact was stored (UPDATE DISPLAY)
```json
{
  "event": "workflow",
  "type": "artifact_set",
  "step": "plan",
  "message": "Artifact 'plan' has been set"
}
```

**`pr_set`** - PR number was set for tracking
```json
{
  "event": "workflow",
  "type": "pr_set",
  "step": "pr",
  "message": "PR #123 set for tracking"
}
```

**`pr_check`** - PR comment check result
```json
{
  "event": "workflow",
  "type": "pr_check",
  "step": "review",
  "message": "No new comments for 1+ minute. Ready for approval."
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

### Option 2: Natural Language
Users can just say:
- "looks good", "approved", "lgtm", "go ahead" → triggers approval
- Any feedback or questions → triggers iteration

### Option 3: Chat Commands
Explicit commands also work:
- `/workflow-approve` - Approve and continue
- `/workflow-iterate <feedback>` - Request changes

### Triggering Approval
Send to Claude (any of these work):
```
looks good
```
```
approved, go ahead
```
```
/workflow-approve
```

### Triggering Iteration
Send to Claude:
```
Can you add error handling to the plan?
```
```
/workflow-iterate Please add error handling
```

## Displaying Artifacts

Access artifact content via `artifacts.<type>.content`:

**Plan** (`artifacts.plan.content`) - markdown string
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
Plan ✓ → Criteria ✓ → Execute ✓ → Verify ✓ → PR ✓ → Review ● → Complete ○
[======================================|     ] 71%
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
interface Artifact {
  type: string;
  content: any;
  step: string;
  created_at: string;
}

interface WorkflowDisplay {
  task: string;
  currentStep: string;
  waitingForApproval: boolean;
  artifacts: Record<string, Artifact>;
  prNumber: number | null;
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
    artifacts: state.artifacts || {},
    prNumber: state.pr_number || null,
    progress: (completed / state.steps.length) * 100,
    steps: state.steps,
  };
}

// Helper to get artifact content
function getArtifact<T>(state: WorkflowDisplay, type: string): T | null {
  return state.artifacts[type]?.content as T || null;
}

// Usage
const plan = getArtifact<string>(state, 'plan');
const criteria = getArtifact<string[]>(state, 'criteria');
const pr = getArtifact<{number: number, url: string}>(state, 'pr');
```

## File Locations

| File | Purpose |
|------|---------|
| `~/state/workflow_state.json` | Current workflow state |
| `workflow.yaml` | Workflow configuration |

## Tips

1. **Poll frequency**: Check state file every 1-2 seconds during active work
2. **Event parsing**: Events appear inline in Claude's output, parse JSON objects with `"event": "workflow"`
3. **Plan display**: Use a collapsible panel - plans can be long
4. **Approval prominence**: Make approval buttons very visible when `waiting_for_approval: true`
5. **Iteration history**: Show `iteration_count` and `iteration_feedback` so users see revision history
