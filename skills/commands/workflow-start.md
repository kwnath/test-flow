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
- **instructions**: What to do in this step

## Default Steps

1. **plan** (approval required) - Explore codebase, design approach, set verification criteria
2. **execute** - Implement changes
3. **verify** - Run tests, check criteria
4. **pr** (approval required) - Create pull request
5. **complete** - Summarize accomplishments

## Setting Verification Criteria

During the plan step, use `workflow_set_criteria` to define what will be verified:

```
workflow_set_criteria(criteria: [
  "npm test passes",
  "No TypeScript errors",
  "Login flow works in browser"
])
```

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
| Step | Approval |
|------|----------|
| ► plan | required |
| ○ execute | - |
| ○ verify | - |
| ○ pr | required |
| ○ complete | - |

### Plan Step Instructions
[Show instructions from workflow response]

Let me explore the codebase to understand the authentication flow...
```

## Approval Gates

Steps with `needs_approval: true` will wait for user feedback. When presenting your plan or PR, wait for the user to approve before calling `workflow_next`.

**Approval phrases to detect:**
- "looks good", "lgtm", "approved", "ship it"
- "yes", "go ahead", "proceed"
- Any positive confirmation

When you detect approval, call `workflow_approve` then `workflow_next` to proceed.
