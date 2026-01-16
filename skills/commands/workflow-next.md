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
3. Show the next step's instructions
4. Begin working on the next step

## Automatic Progression

You should call `workflow_next` when:
- Step work is complete (for non-approval steps)
- User has approved (for approval steps)

### Approval Detection

When a step requires approval and the user says something like:
- "looks good", "lgtm", "approved", "ship it"
- "yes", "go ahead", "proceed", "continue"
- Any clear positive confirmation

**Do this automatically:**
1. Call `workflow_approve` to clear the approval flag
2. Call `workflow_next` to move to the next step
3. Don't ask "should I proceed?" - just proceed

### Example Flow

**Plan step (requires approval):**
```
User: "looks good, proceed"

[Call workflow_approve]
[Call workflow_next]

Completing **plan** step and moving to **execute**.

## Execute Step

Now I'll implement the changes...
```

**Execute step (no approval needed):**
```
[After completing implementation]

Implementation complete. Moving to verification.

[Call workflow_next]

## Verify Step

Running the verification criteria...
```

## When NOT to Use

Don't use `workflow_next` when:
- Blocked by external dependencies (use `workflow_blocked` instead)
- Still working on current step
- Waiting for user approval (wait for their response first)
