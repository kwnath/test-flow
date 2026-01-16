# CLAUDE.md

## Dynamic Workflow System

This project uses a configurable workflow system for structured task management. Workflows are defined in YAML and support approval gates for human checkpoints.

**To start a workflow, use `/workflow-start <task>`**

### Workflow Configuration

Workflows are defined in `workflow.yaml`:

```yaml
name: default
description: Standard development workflow

steps:
  - name: plan
    needs_approval: true
    allows_iteration: true
    approval_prompt: "Review the plan..."
    instructions: |
      Explore the codebase and design your approach...
```

### Available Tools

- `workflow_init(task)` - Initialize workflow, returns first step instructions
- `workflow_status()` - Get current state, progress, plan, and instructions
- `workflow_next()` - Request to proceed. If step requires approval, sets `awaiting_approval` status
- `workflow_approve()` - Approve current step and move to next (only when `awaiting_approval`)
- `workflow_iterate(feedback)` - Provide feedback and iterate on current step
- `workflow_set_plan(plan)` - Store the implementation/design plan for display
- `workflow_set_criteria(criteria[])` - Set verification criteria
- `workflow_step(step, status)` - Update a specific step's status
- `workflow_blocked(reason)` - Mark as blocked by external dependency

### Available Commands

- `/workflow-start <task>` - Initialize workflow
- `/workflow-status` - Show current progress
- `/workflow-next` - Move to next step (or request approval)
- `/workflow-approve` - Approve current step and proceed
- `/workflow-iterate <feedback>` - Provide feedback and iterate
- `/workflow-blocked <reason>` - Mark as blocked

### Default Workflow Steps

1. **plan** (requires approval, allows iteration) - Explore codebase, design approach
2. **criteria** (requires approval, allows iteration) - Define completion criteria
3. **execute** (allows iteration) - Implement changes
4. **verify** (allows iteration) - Run tests, check all criteria pass
5. **pr** (requires approval) - Create pull request
6. **complete** - Summarize accomplishments

### Approval Flow

**CRITICAL: Steps with `requires_approval: true` require human approval before proceeding.**

The approval flow works as follows:

1. You work on a step (e.g., create a plan with diagrams)
2. When done, call `workflow_next()` - this sets status to `awaiting_approval`
3. **STOP AND WAIT** - Do NOT proceed until user responds
4. User reviews and responds with:
   - `/workflow-approve` - Approved, move to next step
   - `/workflow-iterate <feedback>` - Revise based on feedback
5. If iterating, revise your work and call `workflow_next()` again
6. Repeat until approved

**DO NOT auto-proceed through approval gates. Always wait for explicit user approval.**

### Plan Step Instructions

When in the **plan** step:

1. Explore the codebase thoroughly using available tools
2. Design your implementation approach
3. **Include ASCII art or Mermaid diagrams** to visualize:
   - System architecture
   - Data flow
   - Component relationships
   - Before/after states
4. Present the complete plan to the user in a clear, structured format
5. **Call `workflow_set_plan(plan)` to store the COMPLETE plan** - save the ENTIRE plan exactly as you presented it to the user, including all diagrams, options, details, and explanations. Do NOT summarize or condense it.
6. Call `workflow_next()` to request approval
7. **STOP AND WAIT** for the user to respond with either:
   - `/workflow-approve` - Proceed to criteria step
   - `/workflow-iterate <feedback>` - Revise the plan based on feedback
8. If user provides iteration feedback, revise the plan, update with `workflow_set_plan()` (full plan again), and repeat

**DO NOT proceed to the criteria step until the user explicitly approves the plan.**

**IMPORTANT: The plan saved via `workflow_set_plan()` must be the COMPLETE plan, not a summary. External apps display this plan to users, so it must contain all details, diagrams, options, and explanations.**

### Criteria Step Instructions

When in the **criteria** step:

1. Based on the approved plan, define specific completion criteria
2. Criteria should be measurable and verifiable
3. Use `workflow_set_criteria()` to record the criteria
4. Present criteria to the user for review
5. Call `workflow_next()` to request approval
6. **STOP AND WAIT** for user approval or iteration feedback

**DO NOT proceed to execute until the user explicitly approves the criteria.**

### Verification Criteria

During planning/criteria, define criteria to verify later:

```
workflow_set_criteria(criteria: [
  "npm test passes",
  "No TypeScript errors",
  "Login flow works in browser"
])
```

In the verify step, execute each criterion and confirm it passes.

### Event Output

Tools emit structured events for external systems:

```json
{"event": "workflow", "type": "awaiting_approval", "step": "plan", "approval_prompt": "...", "can_iterate": true}
```

```json
{"event": "workflow", "type": "approved", "step": "plan", "next_step": "criteria"}
```

```json
{"event": "workflow", "type": "iteration", "step": "plan", "message": "feedback here"}
```

Event types: `init`, `step_update`, `step_complete`, `awaiting_approval`, `approved`, `iteration`, `blocked`, `criteria_set`

### State Persistence

State is saved to `~/state/workflow_state.json`. After context compaction, call `workflow_status()` to restore awareness.

State includes:
- `iteration_count` - How many times current step has been iterated
- `iteration_feedback` - Array of all feedback received on current step
- `waiting_for_approval` - Whether step is awaiting approval

### Best Practices

1. Always start tasks with `/workflow-start`
2. **Include diagrams** (ASCII art or Mermaid) in plans
3. Set verification criteria during planning
4. **STOP AND WAIT** at approval gates - never auto-proceed
5. Use `workflow_blocked` only for external dependencies (not approval gates)
6. Verify all criteria before creating PR
7. Emit structured events for external parsing
