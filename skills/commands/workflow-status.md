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
3. Show progress, current step, and approval status
4. Display verification criteria if set

## Example Output

```
## Workflow Status

**Task:** Fix authentication bug where users can't login with SSO
**Progress:** 40% (2/5 steps complete)
**Current Step:** execute
**Waiting for Approval:** No

### Verification Criteria
- [ ] npm test passes
- [ ] No TypeScript errors
- [ ] Login flow works in browser

### Steps
| Step | Status | Approval |
|------|--------|----------|
| ✓ plan | completed | required |
| ► execute | in_progress | - |
| ○ verify | pending | - |
| ○ pr | pending | required |
| ○ complete | pending | - |

### Current Step Instructions
Implement the changes according to your plan...
```

## Status Icons

- `✓` - completed
- `►` - in_progress
- `○` - pending
- `⚠` - blocked
- `⏳` - waiting for approval

## Waiting for Approval

When `waiting_for_approval` is true, show this prominently:

```
**Status:** ⏳ Waiting for user approval

The plan has been presented. Waiting for user to approve before proceeding.
Say "looks good" or "approved" to continue.
```
