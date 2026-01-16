# Testing Guide

Comprehensive testing procedures for the workflow MCP.

## Pre-Deployment Testing

### 1. YAML Validation

```bash
# Validate workflow.yaml syntax
python3 -c "import yaml; yaml.safe_load(open('workflow.yaml'))"

# Or with yq
yq eval '.' workflow.yaml
```

### 2. Go Compilation

```bash
cd mcp/workflow

# Build with verbose output
go build -v -o workflow-mcp .

# Run go vet for static analysis
go vet ./...

# Check for race conditions (if tests exist)
go test -race ./...
```

### 3. Dependency Check

```bash
cd mcp/workflow

# Verify dependencies
go mod verify

# Download dependencies
go mod download

# Tidy up go.mod
go mod tidy
```

## Local Testing

### Manual MCP Protocol Testing

Test the MCP server directly via stdin/stdout:

```bash
cd /path/to/project
./mcp/workflow/workflow-mcp << 'EOF'
{"jsonrpc":"2.0","id":1,"method":"initialize"}
EOF
```

Expected response includes `protocolVersion` and `serverInfo`.

### Test Tool Calls

```bash
# List available tools
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./mcp/workflow/workflow-mcp

# Initialize workflow
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"workflow_init","arguments":{"task":"Test task"}}}' | ./mcp/workflow/workflow-mcp

# Get status
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"workflow_status","arguments":{}}}' | ./mcp/workflow/workflow-mcp
```

## Integration Testing

### Test Script

Create `test_workflow.sh`:

```bash
#!/bin/bash
set -e

PROJECT_DIR=$(pwd)
MCP="$PROJECT_DIR/mcp/workflow/workflow-mcp"

echo "Building MCP..."
cd mcp/workflow && go build -o workflow-mcp . && cd ../..

echo "Testing initialize..."
RESP=$(echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | $MCP)
echo "$RESP" | grep -q "protocolVersion" || { echo "FAIL: initialize"; exit 1; }

echo "Testing tools/list..."
RESP=$(echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | $MCP)
echo "$RESP" | grep -q "workflow_init" || { echo "FAIL: tools/list"; exit 1; }
echo "$RESP" | grep -q "workflow_approve" || { echo "FAIL: workflow_approve missing"; exit 1; }
echo "$RESP" | grep -q "workflow_set_criteria" || { echo "FAIL: workflow_set_criteria missing"; exit 1; }

echo "Testing workflow_init..."
RESP=$(echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"workflow_init","arguments":{"task":"Test task"}}}' | $MCP)
echo "$RESP" | grep -q "workflow_id" || { echo "FAIL: workflow_init"; exit 1; }
echo "$RESP" | grep -q "waiting_for_approval" || { echo "FAIL: waiting_for_approval"; exit 1; }

echo "Testing state persistence..."
[ -f tmp/workflow-state.json ] || { echo "FAIL: state not persisted"; exit 1; }

echo "Testing workflow_set_criteria..."
RESP=$(echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"workflow_set_criteria","arguments":{"criteria":["test1","test2"]}}}' | $MCP)
echo "$RESP" | grep -q "criteria_set" || { echo "FAIL: workflow_set_criteria"; exit 1; }

echo "Testing workflow_approve..."
RESP=$(echo '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"workflow_approve","arguments":{}}}' | $MCP)
echo "$RESP" | grep -q "approved" || { echo "FAIL: workflow_approve"; exit 1; }

echo "Testing workflow_next..."
RESP=$(echo '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"workflow_next","arguments":{}}}' | $MCP)
echo "$RESP" | grep -q "current_step" || { echo "FAIL: workflow_next"; exit 1; }

echo ""
echo "All tests passed!"

# Cleanup
rm -f tmp/workflow-state.json
```

### Run Tests

```bash
chmod +x test_workflow.sh
./test_workflow.sh
```

## Unit Tests (Go)

Create `mcp/workflow/workflow_test.go`:

```go
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create temp workflow.yaml
	tmpDir := t.TempDir()
	configFile = filepath.Join(tmpDir, "workflow.yaml")

	yaml := `
name: test
steps:
  - name: step1
    needs_approval: true
    instructions: Do step 1
  - name: step2
    needs_approval: false
    instructions: Do step 2
`
	os.WriteFile(configFile, []byte(yaml), 0644)

	loadConfig()

	if config.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", config.Name)
	}
	if len(config.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(config.Steps))
	}
	if !config.Steps[0].NeedsApproval {
		t.Error("expected step1 to need approval")
	}
}

func TestWorkflowInit(t *testing.T) {
	loadConfig() // Use default config

	result := workflowInit("Test task")

	var resp map[string]any
	json.Unmarshal([]byte(result), &resp)

	if resp["task"] != "Test task" {
		t.Errorf("expected task 'Test task', got '%v'", resp["task"])
	}
	if resp["current_step"] != "plan" {
		t.Errorf("expected current_step 'plan', got '%v'", resp["current_step"])
	}
	if resp["waiting_for_approval"] != true {
		t.Error("expected waiting_for_approval to be true")
	}
}

func TestWorkflowApprove(t *testing.T) {
	loadConfig()
	workflowInit("Test task")

	if !state.WaitingForApproval {
		t.Error("should be waiting for approval initially")
	}

	result := workflowApprove()

	var resp map[string]any
	json.Unmarshal([]byte(result), &resp)

	if resp["approved"] != true {
		t.Error("expected approved to be true")
	}
	if state.WaitingForApproval {
		t.Error("should not be waiting for approval after approve")
	}
}

func TestWorkflowSetCriteria(t *testing.T) {
	loadConfig()
	workflowInit("Test task")

	criteria := []string{"test passes", "no errors"}
	result := workflowSetCriteria(criteria)

	var resp map[string]any
	json.Unmarshal([]byte(result), &resp)

	if resp["criteria_set"] != true {
		t.Error("expected criteria_set to be true")
	}
	if len(state.VerificationCriteria) != 2 {
		t.Errorf("expected 2 criteria, got %d", len(state.VerificationCriteria))
	}
}
```

Run unit tests:

```bash
cd mcp/workflow
go test -v ./...
```

## End-to-End Testing in Claude Code

### Test Sequence

1. Start a workflow:
   ```
   /workflow-start Test the workflow system
   ```

2. Check status:
   ```
   /workflow-status
   ```

3. Test approval detection - say "looks good" and verify Claude proceeds automatically

4. Check state file:
   ```bash
   cat tmp/workflow-state.json
   ```

5. Test blocked state:
   ```
   /workflow-blocked Testing blocked state
   ```

6. Continue workflow:
   ```
   /workflow-next
   ```

## Release Checklist

- [ ] All unit tests pass (`go test ./...`)
- [ ] Integration test script passes
- [ ] YAML validation passes
- [ ] Go vet shows no issues
- [ ] MCP responds to initialize request
- [ ] All 7 tools appear in tools/list
- [ ] workflow_init creates state file
- [ ] workflow_approve clears waiting_for_approval
- [ ] workflow_set_criteria stores criteria
- [ ] State persists across MCP restarts
- [ ] Custom workflow.yaml is loaded correctly
- [ ] Default config works when workflow.yaml is missing

## CI/CD Pipeline Example

```yaml
name: Test Workflow MCP

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Build
        run: |
          cd mcp/workflow
          go build -v -o workflow-mcp .

      - name: Vet
        run: |
          cd mcp/workflow
          go vet ./...

      - name: Test
        run: |
          cd mcp/workflow
          go test -v ./...

      - name: Integration Test
        run: |
          chmod +x test_workflow.sh
          ./test_workflow.sh
```
