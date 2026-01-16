# CLAUDE.md

## CRITICAL: Always Use Workflow First

**When the user asks you to build, create, implement, or fix anything, you MUST immediately call `workflow_init(task)` as your FIRST action.**

Do NOT:
- Ask clarifying questions first
- Ask what kind of feature they want
- Wait for the user to explicitly say `/workflow-start`

Do THIS:
1. User says "Create a todo app" → Immediately call `workflow_init("Create a todo app")`
2. User says "Fix the login bug" → Immediately call `workflow_init("Fix the login bug")`
3. User says "Add dark mode" → Immediately call `workflow_init("Add dark mode")`

The workflow system handles planning and requirements gathering. Start the workflow FIRST, then explore and plan during the plan step.

---

## Dynamic Workflow System

This project uses a configurable workflow system for structured task management. Workflows are defined in YAML and support approval gates for human checkpoints.

**To start a workflow, use `/workflow-start <task>`**

### Workflow Configuration

Workflows are defined in `workflow.yaml`:

```yaml
name: default
description: Standard development workflow

steps:
  - name: plan
    needs_approval: true
    allows_iteration: true
    approval_prompt: "Review the plan..."
    instructions: |
      Explore the codebase and design your approach...
```

### Available Tools

- `workflow_init(task)` - Initialize workflow, returns first step instructions
- `workflow_status()` - Get current state, progress, artifacts, and instructions
- `workflow_next()` - Request to proceed. If step requires approval, sets `awaiting_approval` status
- `workflow_approve()` - Approve current step and move to next (only when `awaiting_approval`)
- `workflow_iterate(feedback)` - Provide feedback and iterate on current step
- `workflow_set_artifact(type, content)` - Store any artifact (plan, criteria, test_results, etc.)
- `workflow_set_plan(plan)` - Store the implementation plan (shorthand for set_artifact)
- `workflow_set_criteria(criteria[])` - Set verification criteria (shorthand for set_artifact)
- `workflow_set_pr(pr_number, pr_url, branch)` - Set PR details for tracking
- `workflow_check_pr(comment_count)` - Check for new PR comments, returns suggested action
- `workflow_add_criteria(criteria[])` - Append criteria from any step (flexible)
- `workflow_goto(step)` - Jump to any step without resetting
- `workflow_step(step, status)` - Update a specific step's status
- `workflow_blocked(reason)` - Mark as blocked by external dependency

### Available Commands

- `/workflow-start <task>` - Initialize workflow
- `/workflow-status` - Show current progress
- `/workflow-next` - Move to next step (or request approval)
- `/workflow-approve` - Approve current step and proceed
- `/workflow-iterate <feedback>` - Provide feedback and iterate
- `/workflow-blocked <reason>` - Mark as blocked

### Default Workflow Steps

1. **plan** (requires approval, allows iteration) - Explore codebase, design approach
2. **criteria** (requires approval, allows iteration) - Define completion criteria
3. **execute** (allows iteration) - Implement changes
4. **verify** (allows iteration) - Run tests, check all criteria pass
5. **pr** - Create pull request, track with `workflow_set_pr()`
6. **review** (auto-loops) - Monitor PR comments, address feedback
7. **human_review** (requires approval) - Notify user PR is ready for their review
8. **complete** - Summarize accomplishments

### Approval Flow

**CRITICAL: Steps with `requires_approval: true` require human approval before proceeding.**

The approval flow works as follows:

1. You work on a step (e.g., create a plan with diagrams)
2. When done, call `workflow_next()` - this sets status to `awaiting_approval`
3. **STOP AND WAIT** - Do NOT proceed until user responds
4. User reviews and responds - detect their intent:
   - **Approval signals** → call `workflow_approve()`:
     - "looks good", "approved", "lgtm", "go ahead", "proceed", "yes", "ok", "ship it"
   - **Iteration signals** → call `workflow_iterate(feedback)`:
     - Any feedback, questions, or change requests
   - **Explicit commands** (also work):
     - `/workflow-approve` - Approved
     - `/workflow-iterate <feedback>` - Iterate with feedback
5. If iterating, revise your work and call `workflow_next()` again
6. Repeat until approved

**DO NOT auto-proceed through approval gates. Always wait for user response.**

### Flexible Navigation

The workflow is **not strictly linear**. You can:

- **Add criteria anytime**: `workflow_add_criteria()` appends from any step
- **Jump to any step**: `workflow_goto(step)` moves without resetting
- **Modify artifacts**: Update plan, criteria, summary from anywhere

Example: User asks to add a criterion during review:
```
User: "Also make sure there are no console.logs"
→ workflow_add_criteria(["- [ ] No console.log statements"])
→ Continue with current step
```

Example: User wants to go back and revise the plan:
```
User: "Actually, let's change the approach"
→ workflow_goto("plan")
→ Revise plan, then continue forward
```

### Plan Step Instructions

When in the **plan** step:

**Phase 1: Clarify (if needed)**
- Review the task requirements
- If anything is ambiguous or unclear, **ask clarifying questions first**
- Wait for answers before proceeding to design
- Skip this if requirements are clear

**Phase 2: Design**
1. Explore the codebase thoroughly using available tools
2. Design your implementation approach
3. **Include Mermaid diagrams** to visualize:
   - System architecture
   - Data flow
   - Component relationships
4. Present the complete plan to the user
5. **Call `workflow_set_plan(plan)`** - save the full plan (no tool output noise)
6. Call `workflow_next()` to request approval
7. **STOP AND WAIT** for user response

**DO NOT proceed to the criteria step until the user explicitly approves the plan.**

**IMPORTANT: The plan saved via `workflow_set_plan()` must be the COMPLETE plan, not a summary. External apps display this plan to users. Include all details, diagrams, options, and explanations - but exclude exploration noise (tool outputs, file listings, command results). Just the clean plan.**

### Plan Formatting

Plans should be **well-structured markdown** that's easy to read:

````markdown
## Implementation Plan

### Overview
Brief summary of what we're building and the approach.

### Architecture
```mermaid
graph TD
    A[CLI] --> B[Command Parser]
    B --> C[TodoService]
    C --> D[Storage]
```

### Key Components
1. **Component A** - What it does
2. **Component B** - What it does

### Files to Create/Modify
- `path/to/file.ts` - Description
- `path/to/other.ts` - Description

### Approach
Step-by-step implementation order.
````

**Guidelines:**
- Use **headers** to organize sections
- Use **Mermaid diagrams** for architecture/flow visualization
- Keep it **scannable** - bullets and short paragraphs
- **No tool output** - no "Running: find...", grep results, etc.

### Criteria Step Instructions

When in the **criteria** step:

1. **Gather context first** - read relevant files to inform criteria:
   - README.md, CONTRIBUTING.md, docs/
   - Existing tests (to match testing patterns)
   - Related code (to understand expected behavior)
   - package.json scripts, CI config
2. Based on plan + context, define **high-level acceptance criteria**
3. Keep to **5-8 items max** - each a meaningful check, not a test case
4. Use `workflow_set_criteria()` to record the criteria
5. Present criteria to the user for review
6. Call `workflow_next()` to request approval
7. **STOP AND WAIT** for user approval or iteration feedback

**DO NOT proceed to execute until the user explicitly approves the criteria.**

### PR Step Instructions

When in the **pr** step:

1. Create PR with `gh pr create` (include summary and test plan)
2. Extract PR number, URL, and branch from output
3. Call `workflow_set_pr(pr_number, pr_url, branch)` to track
4. **Show the PR link to user**: "PR created: <url>"
5. Update summary artifact
6. Call `workflow_next()` to start review monitoring

### Review Step Instructions

When in the **review** step, **loop continuously**:

1. Run `gh pr view <pr_number> --comments --json comments` to get comments
2. Call `workflow_check_pr(comment_count)`
3. Based on `action`:
   - `"address_comments"` → Address the feedback, update summary, loop to step 1
   - `"wait"` → Wait 1 minute, loop to step 1
   - `"ready_for_human_review"` → Call `workflow_next()` to move to human_review

**Stops after 5 mins of no new comments**, then moves to human_review for final approval.

### Human Review Step Instructions

When in the **human_review** step:

1. Present PR to user:
   - PR link
   - Summary of what was implemented
   - Any comments that were addressed
2. Call `workflow_next()` to notify user
3. **STOP AND WAIT** for user to review

This is a "PR is ready for your eyes" notification. User reviews and says "looks good" → complete.

### Verification Criteria Format

Criteria should be **high-level acceptance checks** formatted as a **markdown checklist**:

**GOOD** (high-level checklist with checkboxes):
```
workflow_set_criteria([
  "- [ ] All CRUD operations work correctly",
  "- [ ] Invalid inputs show appropriate errors",
  "- [ ] Help/usage displays when needed",
  "- [ ] No crashes or unhandled exceptions",
  "- [ ] Documentation covers all features"
])
```

**BAD** (too granular - reads like unit tests):
```
workflow_set_criteria([
  "- [ ] node app.js add 'Task' creates task",
  "- [ ] node app.js add '' shows error",
  "- [ ] node app.js list shows tasks",
  "- [ ] node app.js done 1 marks complete",
  // ... 10+ more detailed cases
])
```

**During verify step**, mark items complete as you verify them:
```
workflow_set_criteria([
  "- [x] All CRUD operations work correctly",
  "- [x] Invalid inputs show appropriate errors",
  "- [ ] Help/usage displays when needed",  // in progress
  "- [ ] No crashes or unhandled exceptions",
  "- [ ] Documentation covers all features"
])
```

**Guidelines:**
- **5-8 items max** - group related checks
- **Use `- [ ]` format** - renders as checkboxes
- **Mark `- [x]` when verified** - shows progress
- **Acceptance-level** - "feature works" not "specific command works"

### Event Output

Tools emit structured events for external systems:

```json
{"event": "workflow", "type": "awaiting_approval", "step": "plan", "approval_prompt": "...", "can_iterate": true}
```

```json
{"event": "workflow", "type": "approved", "step": "plan", "next_step": "criteria"}
```

```json
{"event": "workflow", "type": "iteration", "step": "plan", "message": "feedback here"}
```

Event types: `init`, `step_update`, `step_complete`, `awaiting_approval`, `approved`, `iteration`, `blocked`, `artifact_set`, `pr_set`, `pr_check`

### State Persistence

State is saved to `~/state/workflow_state.json`. After context compaction, call `workflow_status()` to restore awareness.

State includes:
- `artifacts` - Map of stored artifacts (plan, criteria, pr, test_results, etc.)
- `iteration_count` - How many times current step has been iterated
- `iteration_feedback` - Array of all feedback received on current step
- `waiting_for_approval` - Whether step is awaiting approval
- `pr_number`, `last_comment_check`, `last_comment_count` - PR tracking for review step

### Summary Artifact

**Keep a context-rich summary** so anyone can pick up the project:

```markdown
**Goal:** Build a CLI todo app with add/list/done/remove commands

**Context:** Node.js CLI storing todos in JSON file with auto-incrementing IDs

**Done:**
- Designed: CLI → Parser → TodoService → JSON storage
- Implemented: add, list, done, remove commands
- Added: input validation and error messages

**Now:** Verifying all criteria (step 4/7)

**Next:** Create PR
```

Store as markdown string via `workflow_set_artifact("summary", content)`.

**Fields:**
- **Goal** - Original task
- **Context** - What it is (tech, approach)
- **Done** - Brief list of completed work
- **Now** - Current step and position
- **Next** - What comes after

**Update at each step transition.** Include enough context for someone with no prior knowledge to pick up the project.

### Best Practices

1. Always start tasks with `/workflow-start`
2. **Include Mermaid diagrams** in plans for architecture/flow
3. Set verification criteria during planning
4. **Update summary artifact** at each step transition
5. **STOP AND WAIT** at approval gates - never auto-proceed
6. Use `workflow_blocked` only for external dependencies (not approval gates)
7. Verify all criteria before creating PR
8. Emit structured events for external parsing
