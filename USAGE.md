# Usage Guide

This guide provides detailed examples and best practices for using the Go No-Nil Linter.

## Table of Contents

- [Quick Start](#quick-start)
- [Command Line Usage](#command-line-usage)
- [Common Patterns](#common-patterns)
- [Integration Examples](#integration-examples)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Quick Start

### 1. Install the Linter

```bash
go install github.com/nickheyer/go_no_nil_linter/cmd/nonillinter@latest
```

### 2. Run on Your Code

```bash
# From your project root
nonillinter ./...
```

### 3. Fix Reported Issues

The linter will report any nil assignments to required message fields. Fix them by ensuring all required message fields are properly initialized.

## Command Line Usage

### Basic Commands

```bash
# Analyze current directory
nonillinter .

# Analyze specific package
nonillinter ./internal/handlers

# Analyze all packages recursively
nonillinter ./...

# Analyze multiple packages
nonillinter ./pkg/... ./cmd/...
```

### Flags

```bash
# Verbose output (shows more details)
nonillinter -v ./...

# Help
nonillinter -h

# Version information
nonillinter -V
```

### Exit Codes

- `0` - No issues found
- `1` - Issues found
- `2` - Analysis error

## Common Patterns

### Pattern 1: Response Builder Functions

**Bad:**

```go
func BuildUserResponse(id string, name string) *UserResponse {
    return &UserResponse{
        User: &User{
            Id:   id,
            Name: name,
            // ❌ Missing required fields: Address, CreatedAt, ContactInfo
        },
        // ❌ Missing required field: LastLogin
    }
}
```

**Good:**

```go
func BuildUserResponse(id string, name string) *UserResponse {
    return &UserResponse{
        User: &User{
            Id:   id,
            Name: name,
            Address: &Address{
                Street:   "Unknown",
                City:     "Unknown",
                Location: &Location{Latitude: 0, Longitude: 0},
            },
            CreatedAt:   timestamppb.Now(),
            ContactInfo: &ContactInfo{Email: "", Phone: ""},
        },
        LastLogin: &date.Date{
            Year:  2024,
            Month: 1,
            Day:   1,
        },
    }
}
```

### Pattern 2: Conditional Initialization

**Bad:**

```go
func GetUser(includeAddress bool) *UserResponse {
    user := &User{
        Id:        "123",
        Name:      "John",
        CreatedAt: timestamppb.Now(),
        ContactInfo: &ContactInfo{
            Email: "john@example.com",
            Phone: "555-1234",
        },
    }
    
    if includeAddress {
        user.Address = &Address{
            Street:   "123 Main St",
            City:     "NYC",
            Location: &Location{Latitude: 40.7128, Longitude: -74.0060},
        }
    }
    // ❌ If includeAddress is false, Address is nil
    
    return &UserResponse{
        User:      user,
        LastLogin: &date.Date{Year: 2024, Month: 1, Day: 1},
    }
}
```

**Good:**

```go
func GetUser(includeAddress bool) *UserResponse {
    var addr *Address
    if includeAddress {
        addr = &Address{
            Street:   "123 Main St",
            City:     "NYC",
            Location: &Location{Latitude: 40.7128, Longitude: -74.0060},
        }
    } else {
        // Always provide a default
        addr = &Address{
            Street:   "",
            City:     "",
            Location: &Location{Latitude: 0, Longitude: 0},
        }
    }
    
    user := &User{
        Id:          "123",
        Name:        "John",
        Address:     addr,
        CreatedAt:   timestamppb.Now(),
        ContactInfo: &ContactInfo{Email: "john@example.com", Phone: "555-1234"},
    }
    
    return &UserResponse{
        User:      user,
        LastLogin: &date.Date{Year: 2024, Month: 1, Day: 1},
    }
}
```

### Pattern 3: Error Handling

**Bad:**

```go
func FetchUser(id string) (*UserResponse, error) {
    addr, err := fetchAddress(id)
    if err != nil {
        return nil, err
    }
    
    // ❌ If addr is nil, this violates the linter
    user := &User{
        Id:          id,
        Name:        "John",
        Address:     addr,
        CreatedAt:   timestamppb.Now(),
        ContactInfo: &ContactInfo{Email: "", Phone: ""},
    }
    
    return &UserResponse{
        User:      user,
        LastLogin: &date.Date{Year: 2024, Month: 1, Day: 1},
    }, nil
}
```

**Good:**

```go
func FetchUser(id string) (*UserResponse, error) {
    addr, err := fetchAddress(id)
    if err != nil {
        return nil, err
    }
    
    // Validate addr is not nil before using
    if addr == nil {
        return nil, errors.New("address is required")
    }
    
    user := &User{
        Id:          id,
        Name:        "John",
        Address:     addr,
        CreatedAt:   timestamppb.Now(),
        ContactInfo: &ContactInfo{Email: "", Phone: ""},
    }
    
    return &UserResponse{
        User:      user,
        LastLogin: &date.Date{Year: 2024, Month: 1, Day: 1},
    }, nil
}
```

### Pattern 4: Using Factory Functions

**Good Practice:**

```go
// Factory function ensures all required fields are set
func NewUser(id, name string) *User {
    return &User{
        Id:   id,
        Name: name,
        Address: &Address{
            Street:   "",
            City:     "",
            Location: &Location{Latitude: 0, Longitude: 0},
        },
        CreatedAt:   timestamppb.Now(),
        ContactInfo: &ContactInfo{Email: "", Phone: ""},
    }
}

// Usage
func BuildResponse() *UserResponse {
    return &UserResponse{
        User:      NewUser("123", "John"),
        LastLogin: &date.Date{Year: 2024, Month: 1, Day: 1},
    }
}
```

### Pattern 5: Optional vs Required

**Understanding the Difference:**

```protobuf
message User {
  string id = 1;                    // Required scalar (linter ignores)
  string name = 2;                  // Required scalar (linter ignores)
  Address address = 3;              // Required message (linter checks) ✓
  optional string nickname = 4;     // Optional scalar (linter ignores)
  optional Address alt_address = 5; // Optional message (linter ignores)
}
```

```go
// ✅ Valid: Optional fields can be nil
user := &User{
    Id:         "123",
    Name:       "John",
    Address:    &Address{...},  // Required: must not be nil
    Nickname:   nil,            // Optional: can be nil
    AltAddress: nil,            // Optional: can be nil
}

// ❌ Invalid: Required field is nil
user := &User{
    Id:      "123",
    Name:    "John",
    Address: nil,  // Error: required field
}
```

## Integration Examples

### Example 1: Makefile Integration

```makefile
.PHONY: lint
lint:
	@echo "Running no-nil linter..."
	@nonillinter ./...

.PHONY: lint-fix
lint-fix:
	@echo "Running no-nil linter with fixes..."
	@nonillinter ./... || (echo "Fix the reported issues" && exit 1)

.PHONY: ci
ci: lint test
	@echo "CI checks passed"
```

### Example 2: GitHub Actions

```yaml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Install linter
        run: go install github.com/nickheyer/go_no_nil_linter/cmd/nonillinter@latest

      - name: Run no-nil linter
        run: nonillinter ./...

      - name: Comment on PR
        if: failure() && github.event_name == 'pull_request'
        uses: actions/github-script@v6
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '❌ No-nil linter found issues. Please fix nil assignments to required protobuf message fields.'
            })
```

### Example 3: Pre-commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash

echo "Running no-nil linter..."

# Run the linter
if ! nonillinter ./...; then
    echo ""
    echo "❌ No-nil linter found issues!"
    echo "Fix the nil assignments to required message fields before committing."
    echo ""
    echo "To bypass this check (not recommended):"
    echo "  git commit --no-verify"
    echo ""
    exit 1
fi

echo "✅ No-nil linter passed"
exit 0
```

Make it executable:

```bash
chmod +x .git/hooks/pre-commit
```

### Example 4: VS Code Integration

Create `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Run No-Nil Linter",
      "type": "shell",
      "command": "nonillinter ./...",
      "problemMatcher": {
        "owner": "go",
        "fileLocation": ["relative", "${workspaceFolder}"],
        "pattern": {
          "regexp": "^(.*):(\\d+):(\\d+):\\s+(.*)$",
          "file": 1,
          "line": 2,
          "column": 3,
          "message": 4
        }
      },
      "group": {
        "kind": "test",
        "isDefault": true
      }
    }
  ]
}
```

Run with `Cmd+Shift+B` (Mac) or `Ctrl+Shift+B` (Linux/Windows).

## Best Practices

### 1. Define Factory Functions

Always create factory functions for complex message types:

```go
func NewAddress(street, city string, lat, lon float64) *Address {
    return &Address{
        Street:     street,
        City:       city,
        PostalCode: "",
        Location: &Location{
            Latitude:  lat,
            Longitude: lon,
        },
    }
}
```

### 2. Use Default Values

For required fields that might not have data, use sensible defaults:

```go
func defaultLocation() *Location {
    return &Location{
        Latitude:  0.0,
        Longitude: 0.0,
    }
}

func defaultAddress() *Address {
    return &Address{
        Street:     "",
        City:       "",
        PostalCode: "",
        Location:   defaultLocation(),
    }
}
```

### 3. Validate at Boundaries

Check for nil at API boundaries:

```go
func CreateUser(req *CreateUserRequest) (*User, error) {
    if req.Address == nil {
        return nil, errors.New("address is required")
    }
    
    return &User{
        Id:          generateID(),
        Name:        req.Name,
        Address:     req.Address,
        CreatedAt:   timestamppb.Now(),
        ContactInfo: req.ContactInfo,
    }, nil
}
```

### 4. Document Optional Fields

Clearly document which fields are optional in proto:

```protobuf
message User {
  string id = 1;
  string name = 2;
  Address address = 3;  // Required: Primary address
  optional Address alternate_address = 4;  // Optional: Secondary address
}
```

### 5. Run Linter in CI

Always run the linter in your CI pipeline to catch issues early.

## Troubleshooting

### Issue: "Linter reports false positives"

**Problem:** The linter reports violations on code that should be valid.

**Solutions:**

1. **Check field optionality**: Ensure the field is not marked as `optional` in proto
2. **Verify message type**: The linter only checks message fields, not scalars
3. **Check generated code**: Ensure protobuf code is properly generated with latest buf

### Issue: "Linter doesn't catch my nil assignment"

**Problem:** You have a nil assignment but the linter doesn't report it.

**Possible Reasons:**

1. **Field is optional**: Check if field has `optional` keyword in proto
2. **Scalar field**: Linter only checks message types, not scalars
3. **Complex dataflow**: The linter may not catch all complex patterns

### Issue: "Too many violations in existing codebase"

**Solutions:**

1. **Gradual adoption**: Start with new code, fix old code incrementally
2. **Use in CI for new PRs only**: Only check changed files
3. **Create helper functions**: Build factory functions for common patterns

```bash
# Check only changed files (in CI)
git diff --name-only origin/main...HEAD | grep '.go$' | xargs nonillinter
```

### Issue: "Performance concerns"

**Solutions:**

1. **Run in parallel**: The analyzer uses Go's built-in concurrency
2. **Cache results**: Most CI systems cache dependencies
3. **Analyze specific packages**: Only run on changed packages

```bash
# Example: Only analyze specific package
nonillinter ./internal/handlers
```

## Advanced Usage

### Custom Analysis Scripts

Create a script to analyze and report:

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/nickheyer/go_no_nil_linter/analyzer"
    "golang.org/x/tools/go/analysis"
    "golang.org/x/tools/go/packages"
)

func main() {
    cfg := &packages.Config{
        Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
    }
    
    pkgs, err := packages.Load(cfg, os.Args[1:]...)
    if err != nil {
        fmt.Fprintf(os.Stderr, "load: %v\n", err)
        os.Exit(1)
    }
    
    // Run analyzer on each package
    for _, pkg := range pkgs {
        pass := &analysis.Pass{
            Analyzer:   analyzer.Analyzer,
            Files:      pkg.Syntax,
            Pkg:        pkg.Types,
            TypesInfo:  pkg.TypesInfo,
            ResultOf:   make(map[*analysis.Analyzer]interface{}),
            Report: func(d analysis.Diagnostic) {
                fmt.Printf("%s: %s\n", pkg.Fset.Position(d.Pos), d.Message)
            },
        }
        
        if _, err := analyzer.Analyzer.Run(pass); err != nil {
            fmt.Fprintf(os.Stderr, "analyze: %v\n", err)
        }
    }
}
```

## Next Steps

1. Review the [Architecture Documentation](ARCHITECTURE.md) for implementation details
2. Check the [README](README.md) for installation and basic usage
3. Explore the [testdata](testdata/) directory for more examples
4. Contribute improvements via GitHub

## Support

If you encounter issues or have questions:

- Open an issue: https://github.com/nickheyer/go_no_nil_linter/issues
- Discussions: https://github.com/nickheyer/go_no_nil_linter/discussions