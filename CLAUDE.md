# CLAUDE.md

## Workflow Protocol

This project uses a structured workflow system for task management. The workflow MCP provides tools for tracking progress through standardized steps.

### Available Tools

- `workflow_init(task)` - Initialize a new workflow
- `workflow_status()` - Get current workflow state
- `workflow_step(step, status)` - Update a step's status
- `workflow_next()` - Complete current step, move to next
- `workflow_blocked(reason)` - Mark workflow as needing human intervention

### Workflow Steps

Every task follows these steps:

1. **plan** - Explore codebase, design approach
2. **criteria** - Define completion criteria (what does "done" look like?)
3. **execute** - Implement changes
4. **verify** - Run tests, check criteria are met
5. **pr** - Create pull request
6. **review** - Address review feedback

### Event Output

When you call workflow tools, emit the event JSON clearly so external systems can parse it:

```json
{"event": "workflow", "type": "step_complete", "step": "plan", "next_step": "criteria"}
```

### State Persistence

Workflow state is saved to `./tmp/workflow-state.json`. After context compaction, call `workflow_status()` to restore awareness of current progress.

### Commands

- `/workflow-start <task>` - Initialize workflow for a task
- `/workflow-status` - Show current progress
- `/workflow-next` - Move to next step
- `/workflow-blocked <reason>` - Mark as blocked

### Best Practices

1. Always start tasks with `/workflow-start`
2. Define clear, measurable criteria before executing
3. Verify all criteria are met before creating PR
4. Use `workflow_blocked` when human input is needed
5. Emit structured events for external parsing
