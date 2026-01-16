# CLAUDE.md

## Dynamic Workflow System

This project uses a configurable workflow system for structured task management. Workflows are defined in YAML and support approval gates for human checkpoints.

### Workflow Configuration

Workflows are defined in `workflow.yaml`:

```yaml
name: default
description: Standard development workflow

steps:
  - name: plan
    needs_approval: true
    instructions: |
      Explore the codebase and design your approach...

  - name: execute
    needs_approval: false
    instructions: |
      Implement the changes...
```

### Available Tools

- `workflow_init(task)` - Initialize workflow, returns first step instructions
- `workflow_status()` - Get current state, progress, and instructions
- `workflow_next()` - Complete current step, move to next
- `workflow_approve()` - Clear approval gate (call before workflow_next)
- `workflow_set_criteria(criteria[])` - Set verification criteria
- `workflow_step(step, status)` - Update a specific step's status
- `workflow_blocked(reason)` - Mark as blocked by external dependency

### Default Workflow Steps

1. **plan** (approval required) - Explore codebase, design approach, set verification criteria
2. **execute** - Implement changes
3. **verify** - Run tests, check all criteria pass
4. **pr** (approval required) - Create pull request
5. **complete** - Summarize accomplishments

### Approval Gates

Steps with `needs_approval: true` wait for user confirmation before proceeding.

**Detecting approval:** When the user says something like:
- "looks good", "lgtm", "approved", "ship it"
- "yes", "go ahead", "proceed"

**Response:** Call `workflow_approve()` then `workflow_next()` to proceed automatically. Don't ask "should I proceed?" - just proceed when approval is detected.

### Verification Criteria

During planning, define criteria to verify later:

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
{"event": "workflow", "type": "step_complete", "step": "plan", "next_step": "execute"}
```

Event types: `init`, `step_update`, `step_complete`, `approved`, `blocked`, `criteria_set`

### State Persistence

State is saved to `./tmp/workflow-state.json`. After context compaction, call `workflow_status()` to restore awareness.

### Commands

- `/workflow-start <task>` - Initialize workflow
- `/workflow-status` - Show current progress
- `/workflow-next` - Move to next step
- `/workflow-blocked <reason>` - Mark as blocked

### Best Practices

1. Always start tasks with `/workflow-start`
2. Set verification criteria during planning
3. Detect approval naturally - don't prompt "should I proceed?"
4. Use `workflow_blocked` only for external dependencies (not approval gates)
5. Verify all criteria before creating PR
6. Emit structured events for external parsing
