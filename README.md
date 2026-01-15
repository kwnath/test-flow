# test-flow

A workflow framework for Claude Code running in Coder workspaces.

## Components

```
test-flow/
├── mcp/workflow/          # MCP server (Go)
│   ├── main.go
│   └── go.mod
├── skills/commands/       # Claude skills
│   ├── workflow-start.md
│   ├── workflow-status.md
│   ├── workflow-next.md
│   └── workflow-blocked.md
├── CLAUDE.md              # Workflow protocol
└── tmp/                   # State storage
```

## Installation

### 1. Build MCP server

```bash
cd mcp/workflow
go build -o workflow-mcp .
```

### 2. Configure Claude Code

Add to your Claude Code MCP config (`.claude/settings.json` or Coder template):

```json
{
  "mcpServers": {
    "workflow": {
      "command": "/path/to/workflow-mcp",
      "args": []
    }
  }
}
```

### 3. Copy skills

```bash
mkdir -p .claude/commands
cp skills/commands/* .claude/commands/
```

## Usage

Start a workflow:
```
/workflow-start Fix the authentication bug
```

Check status:
```
/workflow-status
```

Move to next step:
```
/workflow-next
```

Block for human input:
```
/workflow-blocked Waiting for PR approval
```

## Workflow Steps

1. **plan** - Explore and design
2. **criteria** - Define done
3. **execute** - Implement
4. **verify** - Test
5. **pr** - Create PR
6. **review** - Get feedback

## State

State is persisted to `./tmp/workflow-state.json`.

## Events

The MCP emits structured events that can be parsed from AgentAPI:

```json
{
  "event": "workflow",
  "type": "step_complete",
  "workflow_id": "wf_123",
  "step": "execute",
  "next_step": "verify",
  "timestamp": "2024-01-15T12:00:00Z"
}
```

## Integration with Coder

In your Coder template, add the MCP to the claude-code module:

```hcl
module "claude-code" {
  # ...
  mcp = jsonencode({
    mcpServers = {
      workflow = {
        command = "/path/to/workflow-mcp"
        args    = []
      }
    }
  })
}
```
