# /workflow-approve

Approve the current workflow step and move to the next step.

## Usage
```
/workflow-approve
```

## Instructions

When this command is invoked:

1. Call the `workflow_approve` tool
2. Announce the approval and step transition
3. Begin working on the next step

## Requirements

- Current step must be in `awaiting_approval` status
- If step is not awaiting approval, an error is returned

## When to Use

Use this command to approve a step after reviewing Claude's work:

- After reviewing the implementation **plan** and it looks correct
- After reviewing the completion **criteria** and they're appropriate
- After reviewing the **PR** and it's ready to merge

## Example

```
User: /workflow-approve

Claude:
[Call workflow_approve]

## Plan Approved

Moving from **plan** to **criteria** step.

### Criteria Step Instructions
Now I need to define specific, measurable completion criteria...
```

## Approval Flow

The approval system follows this pattern:

1. Claude works on a step (e.g., creates a plan)
2. Claude calls `workflow_next()` which sets status to `awaiting_approval`
3. Claude **STOPS AND WAITS** for user response
4. User reviews and responds with either:
   - `/workflow-approve` - Approve and proceed
   - `/workflow-iterate <feedback>` - Request changes
5. If approved, Claude proceeds to next step
6. If iterating, Claude revises and calls `workflow_next()` again

## Related Commands

- `/workflow-iterate <feedback>` - Provide feedback and iterate on current step
- `/workflow-status` - Check current approval status
