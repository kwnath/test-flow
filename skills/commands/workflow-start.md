# /workflow-start

Initialize a new workflow for the given task.

## Usage
```
/workflow-start <task description>
```

## Instructions

When this command is invoked:

1. Call the `workflow_init` tool with the task description
2. Announce the workflow has started
3. Begin the **plan** step immediately

## Workflow Steps

The workflow follows these steps in order:

1. **plan** - Explore the codebase, understand the problem, design an approach
2. **criteria** - Define specific, measurable completion criteria
3. **execute** - Implement the changes
4. **verify** - Run tests, verify criteria are met
5. **pr** - Create a pull request
6. **review** - Wait for human review, address feedback

## Example

User: `/workflow-start Fix authentication bug where users can't login with SSO`

You should:
```
I'll initialize a workflow for this task.

[Call workflow_init with task: "Fix authentication bug where users can't login with SSO"]

Workflow initialized. Starting with the **plan** step.

Let me explore the codebase to understand the authentication flow...
```

## Important

- Always emit the workflow event JSON so it can be parsed by external systems
- Update step status as you progress using `workflow_step` or `workflow_next`
- If you need human input, use `workflow_blocked`
