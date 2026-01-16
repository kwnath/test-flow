# /workflow-iterate

Provide feedback and iterate on the current workflow step.

## Usage
```
/workflow-iterate <feedback>
```

## Instructions

When this command is invoked:

1. Call the `workflow_iterate` tool with the feedback
2. Acknowledge the feedback received
3. Revise the current step based on the feedback
4. When revisions are complete, call `workflow_next()` to request approval again

## Requirements

- Current step must allow iteration (`allows_iteration: true`)
- Feedback should describe what needs to change

## When to Use

Use this command when Claude's work needs changes:

- The plan is missing important considerations
- The criteria need adjustment
- The approach should be different
- More detail or clarification is needed

## Example

```
User: /workflow-iterate The plan doesn't address error handling. Please add a section on how errors will be handled.

Claude:
[Call workflow_iterate with feedback: "The plan doesn't address error handling..."]

## Iteration Requested

**Iteration:** 1
**Feedback:** The plan doesn't address error handling. Please add a section on how errors will be handled.

I'll revise the plan to include error handling...

[Revise the plan]

Here's the updated plan with error handling:
...

[Call workflow_next to request approval again]

The revised plan is ready for review. Please approve with `/workflow-approve` or provide more feedback with `/workflow-iterate <feedback>`.
```

## Iteration Tracking

The system tracks:
- **iteration_count**: How many times the step has been iterated
- **iteration_feedback**: Array of all feedback received

This history is preserved until moving to the next step.

## Steps That Allow Iteration

| Step | Allows Iteration |
|------|------------------|
| plan | Yes |
| criteria | Yes |
| execute | Yes |
| verify | Yes |
| pr | No |
| complete | No |

## Related Commands

- `/workflow-approve` - Approve and proceed to next step
- `/workflow-status` - Check iteration count and feedback history
