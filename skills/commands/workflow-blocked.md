# /workflow-blocked

Mark the workflow as blocked and needing human intervention.

## Usage
```
/workflow-blocked <reason>
```

## Instructions

When this command is invoked or when you need human input:

1. Call the `workflow_blocked` tool with the reason
2. Clearly explain what is needed from the human
3. Wait for human response before continuing

## When to Use

Use this when:
- Waiting for PR approval
- Need clarification on requirements
- Encountered an error you can't resolve
- Need access/permissions you don't have
- Design decision requires human input

## Example

```
User: /workflow-blocked Waiting for PR review

You:
[Call workflow_blocked with reason: "Waiting for PR review"]

## Workflow Blocked

**Reason:** Waiting for PR review

The pull request has been created and is awaiting review:
https://github.com/org/repo/pull/123

**Next Steps:**
1. Reviewer approves or requests changes
2. If changes requested, I'll address them
3. Once approved, workflow continues to merge

I'll wait for your signal to continue.
```

## Resuming

When the human provides input or the blocker is resolved, use `/workflow-next` to continue.
