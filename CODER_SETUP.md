# Coder Setup Guide

Deploy the workflow MCP to Coder workspaces.

## Option 1: Add to Coder Template (Recommended)

Add the workflow MCP to your Coder template for automatic installation.

### Terraform Configuration

```hcl
# Build workflow MCP during provisioning
resource "coder_script" "workflow_mcp" {
  agent_id     = coder_agent.main.id
  display_name = "Install Workflow MCP"
  icon         = "/icon/code.svg"
  run_on_start = true
  script       = <<-EOF
    #!/bin/bash
    set -e

    # Clone or update test-flow
    if [ -d "/home/coder/.local/share/test-flow" ]; then
      cd /home/coder/.local/share/test-flow
      git pull
    else
      git clone https://github.com/your-org/test-flow.git /home/coder/.local/share/test-flow
    fi

    # Build MCP server
    cd /home/coder/.local/share/test-flow/mcp/workflow
    go build -o /home/coder/.local/bin/workflow-mcp .

    echo "Workflow MCP installed successfully"
  EOF
}

# Configure Claude Code with MCP
module "claude-code" {
  source   = "registry.coder.com/modules/claude-code/coder"
  version  = "1.0.0"
  agent_id = coder_agent.main.id

  mcp = jsonencode({
    mcpServers = {
      workflow = {
        command = "/home/coder/.local/bin/workflow-mcp"
      }
    }
  })
}
```

### Skills Installation

Add skills to the provisioner script:

```bash
# Install workflow skills
mkdir -p /home/coder/.claude/commands
cp /home/coder/.local/share/test-flow/skills/commands/* /home/coder/.claude/commands/
```

## Option 2: Manual Installation

Install in an existing workspace:

```bash
# Clone repository
git clone https://github.com/your-org/test-flow.git ~/.local/share/test-flow

# Build MCP server
cd ~/.local/share/test-flow/mcp/workflow
go build -o ~/.local/bin/workflow-mcp .

# Install skills
mkdir -p ~/.claude/commands
cp ~/.local/share/test-flow/skills/commands/* ~/.claude/commands/

# Configure MCP (add to ~/.claude/settings.json)
cat >> ~/.claude/settings.json << 'EOF'
{
  "mcpServers": {
    "workflow": {
      "command": "/home/coder/.local/bin/workflow-mcp"
    }
  }
}
EOF
```

## Option 3: Per-Project Installation

For project-specific workflows:

```bash
# In your project directory
cp -r /path/to/test-flow/mcp ./
cp -r /path/to/test-flow/skills ./
cp /path/to/test-flow/workflow.yaml ./

# Build locally
cd mcp/workflow
go build -o ../../.claude/workflow-mcp .

# Configure (project-level .claude/settings.json)
{
  "mcpServers": {
    "workflow": {
      "command": "./.claude/workflow-mcp"
    }
  }
}
```

## Complete Template Example

```hcl
terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
    docker = {
      source = "kreuzwerker/docker"
    }
  }
}

data "coder_workspace" "me" {}
data "coder_workspace_owner" "me" {}

resource "coder_agent" "main" {
  os   = "linux"
  arch = "amd64"
  dir  = "/home/coder"

  startup_script = <<-EOF
    #!/bin/bash

    # Ensure Go is available
    export PATH=$PATH:/usr/local/go/bin

    # Install workflow MCP
    if [ ! -f /home/coder/.local/bin/workflow-mcp ]; then
      mkdir -p /home/coder/.local/bin
      mkdir -p /home/coder/.local/share

      git clone https://github.com/your-org/test-flow.git /home/coder/.local/share/test-flow
      cd /home/coder/.local/share/test-flow/mcp/workflow
      go build -o /home/coder/.local/bin/workflow-mcp .
    fi

    # Install skills
    mkdir -p /home/coder/.claude/commands
    cp /home/coder/.local/share/test-flow/skills/commands/* /home/coder/.claude/commands/
  EOF
}

module "claude-code" {
  source   = "registry.coder.com/modules/claude-code/coder"
  version  = "1.0.0"
  agent_id = coder_agent.main.id

  mcp = jsonencode({
    mcpServers = {
      workflow = {
        command = "/home/coder/.local/bin/workflow-mcp"
      }
    }
  })
}

resource "docker_container" "workspace" {
  name  = "coder-${data.coder_workspace_owner.me.name}-${data.coder_workspace.me.name}"
  image = "codercom/enterprise-base:ubuntu"

  env = [
    "CODER_AGENT_TOKEN=${coder_agent.main.token}",
  ]

  command = ["sh", "-c", coder_agent.main.init_script]
}
```

## Verification

After installation, verify in Claude Code:

```
/workflow-status
```

Should show workflow tools are available, or "no workflow initialized" if working correctly.

## Troubleshooting

### MCP not found

```bash
# Check if binary exists
ls -la ~/.local/bin/workflow-mcp

# Check if Go built successfully
cd ~/.local/share/test-flow/mcp/workflow
go build -v -o ~/.local/bin/workflow-mcp .
```

### Skills not loading

```bash
# Check skills directory
ls -la ~/.claude/commands/

# Restart Claude Code to reload skills
```

### State not persisting

```bash
# Check tmp directory exists and is writable
mkdir -p /path/to/project/tmp
ls -la ~/state/workflow_state.json
```

### Permission errors

```bash
# Ensure binary is executable
chmod +x ~/.local/bin/workflow-mcp
```
