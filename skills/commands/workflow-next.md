# /workflow-next

Complete the current step and move to the next one.

## Usage
```
/workflow-next
```

## Instructions

When this command is invoked:

1. Call the `workflow_next` tool
2. Announce the step transition
3. Begin working on the next step

## Example

```
User: /workflow-next

You:
Completing **plan** step and moving to **criteria**.

[Call workflow_next]

## Criteria Step

Now I need to define completion criteria for this task.

Based on my analysis, the completion criteria are:
1. SSO login flow completes successfully
2. All authentication tests pass
3. No TypeScript errors
4. Error messages are user-friendly

[Call workflow_step with step: "criteria", status: "completed"]

Moving to **execute** step...
```

## Automatic Progression

After completing meaningful work in a step, you should call `/workflow-next` to progress. Don't wait to be asked.

## Blocked State

If you cannot proceed (need human input, waiting for PR review, etc.), use `workflow_blocked` instead of `workflow_next`.
