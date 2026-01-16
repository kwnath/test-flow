# /workflow-blocked

Mark the workflow as blocked due to external dependencies.

## Usage
```
/workflow-blocked <reason>
```

## Instructions

When this command is invoked or when you encounter external blockers:

1. Call the `workflow_blocked` tool with the reason
2. Clearly explain what external dependency is blocking progress
3. Wait for the blocker to be resolved

## When to Use

Use `workflow_blocked` for **external dependencies** only:
- Waiting for CI/CD pipeline
- Need access/permissions you don't have
- Waiting for external API or service
- Infrastructure issues
- Waiting for another team's work

## When NOT to Use

**Do NOT use for approval gates.** Steps that require approval use the built-in approval mechanism:
- Present your work to the user
- Wait for approval phrases ("looks good", "approved", etc.)
- Call `workflow_approve` then `workflow_next`

## Difference: Blocked vs Approval

| Situation | Action |
|-----------|--------|
| User needs to review plan | Wait for approval (don't use blocked) |
| User needs to review PR | Wait for approval (don't use blocked) |
| CI pipeline is failing | Use `workflow_blocked` |
| Need database access | Use `workflow_blocked` |
| External service is down | Use `workflow_blocked` |

## Example

```
User: /workflow-blocked Waiting for CI pipeline to complete

You:
[Call workflow_blocked with reason: "Waiting for CI pipeline to complete"]

## Workflow Blocked

**Step:** verify
**Reason:** Waiting for CI pipeline to complete

The CI pipeline is running. Once it completes, I'll continue with verification.

**To unblock:** Run `/workflow-next` when the CI pipeline passes.
```

## Resuming

When the external dependency is resolved:
1. User can run `/workflow-next` to continue
2. Or you can detect the resolution and call `workflow_next` directly
