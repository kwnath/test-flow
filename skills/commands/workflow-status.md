# /workflow-status

Display the current workflow status and progress.

## Usage
```
/workflow-status
```

## Instructions

When this command is invoked:

1. Call the `workflow_status` tool
2. Display the results in a clear, readable format
3. Show progress percentage and current step

## Example Output

```
## Workflow Status

**Task:** Fix authentication bug where users can't login with SSO
**Progress:** 50% (3/6 steps complete)

| Step | Status |
|------|--------|
| ✓ plan | completed |
| ✓ criteria | completed |
| ✓ execute | completed |
| ► verify | in_progress |
| ○ pr | pending |
| ○ review | pending |

**Current Step:** verify - Running tests to verify the fix
```

## Status Icons

- `✓` - completed
- `►` - in_progress
- `○` - pending
- `⚠` - blocked
