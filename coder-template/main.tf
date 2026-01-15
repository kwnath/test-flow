terraform {
  required_providers {
    coder = {
      source  = "coder/coder"
      version = ">= 2.12.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.26.0"
    }
  }
}

provider "coder" {}

provider "kubernetes" {}

data "coder_provisioner" "me" {}
data "coder_workspace" "me" {}
data "coder_workspace_owner" "me" {}

# Task support - enables AI Tasks feature
data "coder_task" "me" {}

# External auth for GitHub (users authorize via OAuth)
data "coder_external_auth" "github" {
  id = "github"
}

# AI Task resource - links task to Claude Code app
resource "coder_ai_task" "task" {
  count  = data.coder_workspace.me.start_count
  app_id = module.claude-code[count.index].task_app_id
}

variable "namespace" {
  type        = string
  default     = "coder-workspaces"
  description = "Kubernetes namespace for workspaces"
}

variable "docker_image" {
  type        = string
  default     = "codercom/enterprise-base:ubuntu"
  description = "Docker image for workspaces"
}

variable "anthropic_api_key" {
  type        = string
  default     = ""
  sensitive   = true
  description = "Anthropic API key for Claude Code"
}

variable "github_repo" {
  type        = string
  default     = "kwnath/test-flow"
  description = "GitHub repo to clone (owner/repo format)"
}

data "coder_parameter" "cpu" {
  name         = "cpu"
  display_name = "CPU Cores"
  type         = "number"
  default      = "2"
  mutable      = true
  option {
    name  = "2 Cores"
    value = "2"
  }
  option {
    name  = "4 Cores"
    value = "4"
  }
}

data "coder_parameter" "memory" {
  name         = "memory"
  display_name = "Memory (GB)"
  type         = "number"
  default      = "4"
  mutable      = true
  option {
    name  = "4 GB"
    value = "4"
  }
  option {
    name  = "8 GB"
    value = "8"
  }
}

data "coder_parameter" "home_disk_size" {
  name         = "home_disk_size"
  display_name = "Home Disk Size (GB)"
  type         = "number"
  default      = "50"
  mutable      = false
  validation {
    min = 10
    max = 100
  }
}

resource "coder_agent" "main" {
  arch = data.coder_provisioner.me.arch
  os   = "linux"

  startup_script = <<-EOT
    set -e

    # Setup GitHub auth using Coder's external auth
    if [ -n "${data.coder_external_auth.github.access_token}" ]; then
      echo "${data.coder_external_auth.github.access_token}" | gh auth login --with-token 2>/dev/null || true
      echo "GitHub auth configured"
    fi

    # Clone or update repo
    mkdir -p ~/projects
    REPO_NAME=$(echo "${var.github_repo}" | cut -d'/' -f2)
    if [ ! -d ~/projects/$REPO_NAME/.git ]; then
      gh repo clone ${var.github_repo} ~/projects/$REPO_NAME -- --depth=1
      echo "Repository cloned"
    else
      cd ~/projects/$REPO_NAME
      git pull --ff-only || true
      echo "Repository updated"
    fi
  EOT

  env = {
    GIT_AUTHOR_NAME     = coalesce(data.coder_workspace_owner.me.full_name, data.coder_workspace_owner.me.name)
    GIT_AUTHOR_EMAIL    = data.coder_workspace_owner.me.email
    GIT_COMMITTER_NAME  = coalesce(data.coder_workspace_owner.me.full_name, data.coder_workspace_owner.me.name)
    GIT_COMMITTER_EMAIL = data.coder_workspace_owner.me.email
    GITHUB_TOKEN        = data.coder_external_auth.github.access_token
  }

  metadata {
    display_name = "CPU Usage"
    key          = "cpu"
    script       = "coder stat cpu"
    interval     = 10
    timeout      = 1
  }

  metadata {
    display_name = "Memory Usage"
    key          = "mem"
    script       = "coder stat mem"
    interval     = 10
    timeout      = 1
  }

  metadata {
    display_name = "Disk Usage"
    key          = "disk"
    script       = "coder stat disk --path $HOME"
    interval     = 60
    timeout      = 1
  }
}

# Workspace setup script - builds workflow MCP
resource "coder_script" "workspace_setup" {
  agent_id           = coder_agent.main.id
  display_name       = "Workspace Setup"
  icon               = "/icon/widgets.svg"
  run_on_start       = true
  start_blocks_login = true
  timeout            = 300

  script = <<-EOT
    set -e

    # Setup GitHub auth
    echo "=== Setting up GitHub auth ==="
    if [ -n "$GITHUB_TOKEN" ]; then
      echo "$GITHUB_TOKEN" | gh auth login --with-token 2>/dev/null || true
      echo "GitHub auth configured"
    fi

    # Clone repo
    echo "=== Cloning repository ==="
    REPO_NAME=$(echo "${var.github_repo}" | cut -d'/' -f2)
    mkdir -p /home/coder/projects
    if [ ! -d /home/coder/projects/$REPO_NAME/.git ]; then
      rm -rf /home/coder/projects/$REPO_NAME
      gh repo clone ${var.github_repo} /home/coder/projects/$REPO_NAME -- --depth=1 --single-branch
      echo "Repository cloned"
    else
      cd /home/coder/projects/$REPO_NAME
      git pull --ff-only || true
      echo "Repository updated"
    fi

    # Build workflow MCP
    echo "=== Building workflow MCP ==="
    cd /home/coder/projects/$REPO_NAME/mcp/workflow
    go build -o /home/coder/go/bin/workflow-mcp . || echo "MCP build failed"

    # Copy skills to .claude/commands
    echo "=== Setting up Claude skills ==="
    mkdir -p /home/coder/projects/$REPO_NAME/.claude/commands
    cp /home/coder/projects/$REPO_NAME/skills/commands/* /home/coder/projects/$REPO_NAME/.claude/commands/ 2>/dev/null || true

    # Create tmp directory for workflow state
    mkdir -p /home/coder/projects/$REPO_NAME/tmp

    echo "=== Workspace setup complete ==="
  EOT
}

# Code server for VS Code in browser
module "code-server" {
  count    = data.coder_workspace.me.start_count
  source   = "registry.coder.com/coder/code-server/coder"
  version  = "~> 1.0"
  agent_id = coder_agent.main.id
  folder   = "/home/coder/projects/test-flow"
  order    = 2
}

# Claude Code module for AI Tasks
module "claude-code" {
  count   = data.coder_workspace.me.start_count
  source  = "registry.coder.com/coder/claude-code/coder"
  version = "4.3.0"

  agent_id                     = coder_agent.main.id
  workdir                      = "/home/coder/projects/test-flow"
  order                        = 999
  model                        = "sonnet"
  claude_api_key               = var.anthropic_api_key
  dangerously_skip_permissions = true
  claude_md_path               = "/home/coder/projects/test-flow/CLAUDE.md"

  # Task prompt from Coder Tasks
  ai_prompt = data.coder_task.me.prompt

  system_prompt = <<-EOT
    You are testing the workflow framework.

    Available workflow commands:
    - /workflow-start <task> - Initialize a new workflow
    - /workflow-status - Check current progress
    - /workflow-next - Move to next step
    - /workflow-blocked <reason> - Mark as needing human input

    Always use the workflow MCP tools to track progress.
    Emit structured JSON events for external parsing.

    Git workflow:
    - Create feature branches: git checkout -b ${data.coder_workspace_owner.me.name}/<feature-name>
  EOT

  mcp = <<-EOF
  {
    "mcpServers": {
      "workflow": {
        "command": "/home/coder/go/bin/workflow-mcp",
        "args": []
      }
    }
  }
  EOF
}

# Tmux module for terminal UI
module "tmux" {
  count    = data.coder_workspace.me.start_count
  source   = "registry.coder.com/anomaly/tmux/coder"
  version  = "~> 1.0"
  agent_id = coder_agent.main.id
  order    = 1
}

# Persistent volume for home directory
resource "kubernetes_persistent_volume_claim_v1" "home" {
  metadata {
    name      = "coder-${data.coder_workspace.me.id}-home"
    namespace = var.namespace
    labels = {
      "app.kubernetes.io/name"     = "coder-pvc"
      "app.kubernetes.io/instance" = "coder-pvc-${data.coder_workspace.me.id}"
      "app.kubernetes.io/part-of"  = "coder"
      "com.coder.resource"         = "true"
      "com.coder.workspace.id"     = data.coder_workspace.me.id
      "com.coder.workspace.name"   = data.coder_workspace.me.name
      "com.coder.user.id"          = data.coder_workspace_owner.me.id
      "com.coder.user.username"    = data.coder_workspace_owner.me.name
    }
  }
  wait_until_bound = false
  spec {
    access_modes = ["ReadWriteOnce"]
    resources {
      requests = {
        storage = "${data.coder_parameter.home_disk_size.value}Gi"
      }
    }
  }
}

# Workspace deployment
resource "kubernetes_deployment_v1" "workspace" {
  count = data.coder_workspace.me.start_count
  depends_on = [
    kubernetes_persistent_volume_claim_v1.home
  ]
  wait_for_rollout = false

  metadata {
    name      = "coder-${data.coder_workspace.me.id}"
    namespace = var.namespace
    labels = {
      "app.kubernetes.io/name"     = "coder-workspace"
      "app.kubernetes.io/instance" = "coder-workspace-${data.coder_workspace.me.id}"
      "app.kubernetes.io/part-of"  = "coder"
      "com.coder.resource"         = "true"
      "com.coder.workspace.id"     = data.coder_workspace.me.id
      "com.coder.workspace.name"   = data.coder_workspace.me.name
      "com.coder.user.id"          = data.coder_workspace_owner.me.id
      "com.coder.user.username"    = data.coder_workspace_owner.me.name
    }
  }

  spec {
    replicas = 1
    selector {
      match_labels = {
        "com.coder.workspace.id" = data.coder_workspace.me.id
      }
    }
    strategy {
      type = "Recreate"
    }

    template {
      metadata {
        labels = {
          "app.kubernetes.io/name"     = "coder-workspace"
          "app.kubernetes.io/instance" = "coder-workspace-${data.coder_workspace.me.id}"
          "app.kubernetes.io/part-of"  = "coder"
          "com.coder.resource"         = "true"
          "com.coder.workspace.id"     = data.coder_workspace.me.id
          "com.coder.workspace.name"   = data.coder_workspace.me.name
          "com.coder.user.id"          = data.coder_workspace_owner.me.id
          "com.coder.user.username"    = data.coder_workspace_owner.me.name
        }
        annotations = {
          "cluster-autoscaler.kubernetes.io/safe-to-evict" = "false"
        }
      }

      spec {
        priority_class_name = "coder-workspace"

        security_context {
          run_as_user = 1000
          fs_group    = 1000
        }

        container {
          name              = "dev"
          image             = var.docker_image
          image_pull_policy = "Always"
          command           = ["sh", "-c", coder_agent.main.init_script]

          security_context {
            run_as_user = 1000
          }

          env {
            name  = "CODER_AGENT_TOKEN"
            value = coder_agent.main.token
          }

          resources {
            requests = {
              cpu    = "250m"
              memory = "512Mi"
            }
            limits = {
              cpu    = "${data.coder_parameter.cpu.value}"
              memory = "${data.coder_parameter.memory.value}Gi"
            }
          }

          volume_mount {
            mount_path = "/home/coder"
            name       = "home"
            read_only  = false
          }
        }

        volume {
          name = "home"
          persistent_volume_claim {
            claim_name = kubernetes_persistent_volume_claim_v1.home.metadata.0.name
            read_only  = false
          }
        }

        affinity {
          pod_anti_affinity {
            preferred_during_scheduling_ignored_during_execution {
              weight = 1
              pod_affinity_term {
                topology_key = "kubernetes.io/hostname"
                label_selector {
                  match_expressions {
                    key      = "app.kubernetes.io/name"
                    operator = "In"
                    values   = ["coder-workspace"]
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}
