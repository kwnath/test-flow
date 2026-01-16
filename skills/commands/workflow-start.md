# /workflow-start

Initialize a new workflow for the given task.

## Usage
```
/workflow-start <task description>
```

## Instructions

When this command is invoked:

1. Call the `workflow_init` tool with the task description
2. Display the workflow configuration (steps with approval gates)
3. Show the current step instructions
4. Begin working on the **plan** step

## Dynamic Workflow

The workflow is loaded from `workflow.yaml` in the project root. Each step has:
- **name**: Step identifier
- **needs_approval**: Whether user approval is required before proceeding
- **allows_iteration**: Whether the step can be iterated with feedback
- **approval_prompt**: Message shown when awaiting approval
- **instructions**: What to do in this step

## Default Steps

1. **plan** (requires approval, allows iteration) - Explore codebase, design approach with diagrams
2. **criteria** (requires approval, allows iteration) - Define completion criteria
3. **execute** (allows iteration) - Implement changes
4. **verify** (allows iteration) - Run tests, check criteria
5. **pr** (requires approval) - Create pull request
6. **complete** - Summarize accomplishments

## Plan Step Requirements

When executing the **plan** step:

1. Explore the codebase thoroughly using available tools
2. Design your implementation approach
3. **Include ASCII art or Mermaid diagrams** to visualize:
   - System architecture
   - Data flow
   - Component relationships
   - Before/after states
4. Present the complete plan to the user in a clear, structured format
5. **Call `workflow_set_plan(plan)` with the COMPLETE plan** - include EVERYTHING you presented: all diagrams, all options, all details, all explanations. Do NOT summarize. External apps display this to users.
6. Call `workflow_next()` to request approval
7. **STOP AND WAIT** - Do not proceed until user approves or provides iteration feedback

The user will respond with:
- `/workflow-approve` - Approved, move to criteria
- `/workflow-iterate <feedback>` - Revise plan based on feedback

## Example

User: `/workflow-start Fix authentication bug where users can't login with SSO`

You should:
```
I'll initialize a workflow for this task.

[Call workflow_init with task: "Fix authentication bug where users can't login with SSO"]

## Workflow Started

**Task:** Fix authentication bug where users can't login with SSO
**Current Step:** plan (requires approval)

### Steps Overview
| Step | Approval | Iteration |
|------|----------|-----------|
| ► plan | required | allowed |
| ○ criteria | required | allowed |
| ○ execute | - | allowed |
| ○ verify | - | allowed |
| ○ pr | required | - |
| ○ complete | - | - |

Now let me explore the codebase to understand the authentication flow...

[Explore codebase, design approach]

## Implementation Plan

### Current Architecture
```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│   Auth API  │────▶│  SSO Provider│
└─────────────┘     └─────────────┘     └─────────────┘
```

### Proposed Changes
1. Fix the token validation in auth.ts
2. Update the SSO callback handler
3. Add error handling for edge cases

### Files to Modify
- src/auth/auth.ts
- src/auth/sso-handler.ts

[Call workflow_next to request approval]

---

**Plan ready for review.** Please:
- `/workflow-approve` to proceed to defining criteria
- `/workflow-iterate <feedback>` to request changes to the plan
```

## Approval Gates

Steps with `requires_approval: true` will enter `awaiting_approval` status when you call `workflow_next()`.

**CRITICAL: You must STOP AND WAIT for user approval. Do NOT automatically proceed.**

When awaiting approval:
1. Present your work clearly
2. Call `workflow_next()` to trigger the approval request
3. **STOP** - Do not continue until user responds
4. User will call `/workflow-approve` or `/workflow-iterate <feedback>`
5. Only then should you proceed or revise

## Iteration

If the user calls `/workflow-iterate <feedback>`:
1. Acknowledge the feedback
2. Revise your work based on the feedback
3. Present the revised work
4. Call `workflow_next()` to request approval again
5. **STOP AND WAIT** again

The system tracks iteration count and all feedback received.
