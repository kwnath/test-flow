# test-flow

A dynamic workflow framework for Claude Code with YAML configuration and approval gates.

## Features

- **YAML Configuration**: Define custom workflows in `workflow.yaml`
- **Approval Gates**: Steps can require user approval before proceeding
- **Natural Approval Detection**: Claude detects phrases like "looks good" and proceeds automatically
- **Verification Criteria**: Define tests/checks during planning, execute during verification
- **State Persistence**: Workflow state survives context resets
- **Structured Events**: JSON events for external system integration

## Quick Start

```bash
# Build the MCP server
cd mcp/workflow
go build -o workflow-mcp .

# Configure Claude Code (add to .claude/settings.json)
{
  "mcpServers": {
    "workflow": {
      "command": "/path/to/workflow-mcp"
    }
  }
}

# Copy skills
mkdir -p .claude/commands
cp skills/commands/* .claude/commands/
```

## Usage

Start a workflow:
```
/workflow-start Fix the authentication bug
```

Claude will:
1. Initialize the workflow
2. Present the plan (waits for approval)
3. Execute changes
4. Verify criteria
5. Create PR (waits for approval)
6. Complete

Approve with natural language:
```
User: looks good, proceed
Claude: [automatically calls workflow_approve and workflow_next]
```

## Workflow Configuration

Create `workflow.yaml` in your project root:

```yaml
name: my-workflow
description: Custom development workflow

steps:
  - name: plan
    needs_approval: true
    instructions: |
      Design your approach and define verification criteria.

  - name: execute
    needs_approval: false
    instructions: |
      Implement the changes.

  - name: verify
    needs_approval: false
    instructions: |
      Run tests and checks.

  - name: pr
    needs_approval: true
    instructions: |
      Create a pull request.

  - name: complete
    needs_approval: false
    instructions: |
      Summarize what was done.
```

## Example Workflows

### Hotfix Workflow

```yaml
name: hotfix
description: Fast-track fixes for production issues

steps:
  - name: diagnose
    needs_approval: false
    instructions: Find the root cause quickly.

  - name: fix
    needs_approval: false
    instructions: Apply minimal fix.

  - name: verify
    needs_approval: false
    instructions: Confirm fix works.

  - name: deploy
    needs_approval: true
    instructions: Deploy to production (requires approval).
```

### Feature Development

```yaml
name: feature
description: Full feature development with design review

steps:
  - name: design
    needs_approval: true
    instructions: Create design doc for review.

  - name: implement
    needs_approval: false
    instructions: Build the feature.

  - name: test
    needs_approval: false
    instructions: Write and run tests.

  - name: review
    needs_approval: true
    instructions: Submit for code review.

  - name: merge
    needs_approval: false
    instructions: Merge when approved.
```

## State Persistence

Workflow state is saved to `~/state/workflow_state.json`:

```json
{
  "id": "wf_1234567890",
  "task": "Fix authentication bug",
  "current_step": "execute",
  "waiting_for_approval": false,
  "verification_criteria": ["npm test passes", "No TS errors"],
  "steps": [...]
}
```

## Events

The MCP emits structured events for external integration:

```json
{
  "event": "workflow",
  "type": "step_complete",
  "workflow_id": "wf_1234567890",
  "step": "plan",
  "next_step": "execute",
  "timestamp": "2024-01-15T12:00:00Z"
}
```

Event types: `init`, `step_update`, `step_complete`, `approved`, `blocked`, `criteria_set`

## Coder Integration

Add to your Coder template:

```hcl
module "claude-code" {
  source = "..."

  mcp = jsonencode({
    mcpServers = {
      workflow = {
        command = "/home/coder/.local/bin/workflow-mcp"
      }
    }
  })
}
```

See [CODER_SETUP.md](CODER_SETUP.md) for detailed deployment instructions.

## Testing

See [TESTING.md](TESTING.md) for testing procedures and validation.

## Project Structure

```
test-flow/
├── workflow.yaml           # Workflow configuration
├── mcp/workflow/           # MCP server (Go)
│   ├── main.go
│   └── go.mod
├── skills/commands/        # Claude skills
│   ├── workflow-start.md
│   ├── workflow-status.md
│   ├── workflow-next.md
│   └── workflow-blocked.md
├── CLAUDE.md               # Protocol documentation
├── CODER_SETUP.md          # Deployment guide
├── TESTING.md              # Testing guide
└── tmp/                    # State storage
```
