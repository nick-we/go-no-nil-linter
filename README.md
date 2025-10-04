# Go No-Nil Linter for Protobuf Messages

A Go static analysis tool that detects nil assignments to non-optional protobuf message fields. This linter helps prevent runtime panics and ensures data integrity in services using protobuf-generated code.

## Overview

This linter specifically targets **message-type fields** (custom messages and Google well-known types) in protobuf-generated Go code, ignoring scalar types. It performs **recursive validation** to ensure that all required message fields throughout the object graph are properly initialized.

### What It Checks

✅ **Custom message fields** - Your own protobuf message types  
✅ **Google well-known types** - `google.protobuf.Timestamp`, `google.type.Date`, etc.  
✅ **Nested message fields** - Recursively validates all submessages  
✅ **Explicit nil assignments** - Direct `field = nil` assignments  
✅ **Implicit nil assignments** - Assignments from nil variables  
✅ **Uninitialized fields** - Required fields not set in composite literals  

### What It Ignores

❌ **Scalar fields** - `string`, `int32`, `bool`, `bytes`, etc.  
❌ **Optional fields** - Fields marked with `optional` keyword  
❌ **Scalar wrappers** - `StringValue`, `Int32Value`, etc.  

## Installation

```bash
go install github.com/nickheyer/go_no_nil_linter/cmd/nonillinter@latest
```

Or build from source:

```bash
git clone https://github.com/nickheyer/go_no_nil_linter.git
cd go_no_nil_linter
go build -o nonillinter ./cmd/nonillinter
```

## Usage

### As a Standalone Tool

Run the linter on your package:

```bash
# Analyze current package
nonillinter .

# Analyze specific package
nonillinter ./path/to/package

# Analyze all packages recursively
nonillinter ./...

# With verbose output
nonillinter -v ./...
```

### As a Library

Use the analyzer in your own tools:

```go
package main

import (
    "github.com/nickheyer/go_no_nil_linter/analyzer"
    "golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
    singlechecker.Main(analyzer.Analyzer)
}
```

### Integration with CI/CD

#### GitHub Actions

```yaml
name: Lint

on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install linter
        run: go install github.com/nickheyer/go_no_nil_linter/cmd/nonillinter@latest
      
      - name: Run linter
        run: nonillinter ./...
```

#### Pre-commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/sh
nonillinter ./...
if [ $? -ne 0 ]; then
    echo "Linter found issues. Commit aborted."
    exit 1
fi
```

## Examples

### Example Protobuf Definition

```protobuf
syntax = "proto3";

package example.v1;

import "google/protobuf/timestamp.proto";
import "google/type/date.proto";

message Address {
  string street = 1;
  string city = 2;
  Location location = 3;  // Required message field
}

message Location {
  double latitude = 1;
  double longitude = 2;
}

message User {
  string id = 1;
  string name = 2;
  Address address = 3;  // Required message field
  google.protobuf.Timestamp created_at = 4;  // Required well-known type
  optional string nickname = 5;  // Optional - can be nil
}

message UserResponse {
  User user = 1;  // Required message field
  google.type.Date last_login = 2;  // Required well-known type
}
```

### ❌ Invalid Code (Will Trigger Violations)

#### Direct Nil Assignment

```go
response := &UserResponse{}
response.User = nil  // ❌ ERROR: nil assignment to non-optional message field 'User'
```

#### Implicit Nil Assignment

```go
var user *User  // user is nil
response := &UserResponse{
    User: user,  // ❌ ERROR: variable 'user' used for field 'User' is nil
}
```

#### Missing Required Field

```go
response := &UserResponse{
    // ❌ ERROR: non-optional message field 'User' not initialized
    // ❌ ERROR: non-optional message field 'LastLogin' not initialized
}
```

#### Nil in Nested Message

```go
user := &User{
    Id: "123",
    Name: "John",
    Address: nil,  // ❌ ERROR: nil assignment to non-optional message field 'Address'
}
```

#### Deep Nesting Violation

```go
addr := &Address{
    Street: "123 Main St",
    City: "NYC",
    Location: nil,  // ❌ ERROR detected through recursive validation
}

user := &User{
    Id: "123",
    Name: "John",
    Address: addr,  // Triggers recursive check
}

response := &UserResponse{
    User: user,  // ❌ ERROR: non-optional message field 'User.Address.Location' not initialized
}
```

### ✅ Valid Code (No Violations)

#### Fully Initialized

```go
response := &UserResponse{
    User: &User{
        Id: "123",
        Name: "John Doe",
        Address: &Address{
            Street: "123 Main St",
            City: "New York",
            Location: &Location{
                Latitude: 40.7128,
                Longitude: -74.0060,
            },
        },
        CreatedAt: timestamppb.Now(),
    },
    LastLogin: &date.Date{
        Year: 2024,
        Month: 1,
        Day: 15,
    },
}
```

#### Optional Fields Can Be Nil

```go
user := &User{
    Id: "123",
    Name: "John",
    Nickname: nil,  // ✅ OK: optional field
    Address: &Address{
        Street: "123 Main St",
        City: "NYC",
        Location: &Location{
            Latitude: 40.7128,
            Longitude: -74.0060,
        },
    },
    CreatedAt: timestamppb.Now(),
}
```

#### Scalar Fields Can Be Zero Values

```go
user := &User{
    Id: "",      // ✅ OK: scalar field (string)
    Name: "",    // ✅ OK: scalar field (string)
    Address: &Address{
        Street: "",
        City: "",
        Location: &Location{
            Latitude: 0.0,   // ✅ OK: scalar field (double)
            Longitude: 0.0,  // ✅ OK: scalar field (double)
        },
    },
    CreatedAt: timestamppb.Now(),
}
```

## Error Messages

The linter provides clear, actionable error messages:

```
user_handler.go:15:2: nil assignment to non-optional message field 'User' in protobuf message 'UserResponse'

user_handler.go:23:5: nil assignment to non-optional message field 'Address' in protobuf message 'User'

user_handler.go:30:3: non-optional message field 'Location' not initialized in protobuf message 'Address'

user_handler.go:40:2: nil assignment to non-optional message field 'User.Address.Location' in protobuf message 'UserResponse'
```

## How It Works

The linter uses Go's static analysis framework to:

1. **Identify protobuf message types** by checking for the `ProtoMessage()` method
2. **Filter message fields** (excluding scalars and optional fields)
3. **Detect nil assignments** (both explicit and implicit)
4. **Recursively validate** nested message structures
5. **Report violations** with precise location and context

### Detection Algorithm

```
For each assignment or initialization:
  1. Check if target is a protobuf message field
  2. Check if field is required (not optional)
  3. If value is nil → Report violation
  4. If value is a message type → Recursively validate all submessages
  5. If field is uninitialized → Report violation
```

## Architecture

```
go_no_nil_linter/
├── analyzer/           # Core analysis logic
│   ├── analyzer.go     # Main analyzer
│   ├── messages.go     # Message type detection
│   └── detector.go     # Nil detection & recursive validation
├── cmd/
│   └── nonillinter/    # CLI tool
├── proto/              # Example protobuf definitions
├── gen/                # Generated Go code
└── testdata/           # Test cases
```

For detailed architecture information, see [`ARCHITECTURE.md`](ARCHITECTURE.md:1).

## Development

### Prerequisites

- Go 1.21 or higher
- Buf CLI for protobuf generation

### Setup

```bash
# Clone repository
git clone https://github.com/nickheyer/go_no_nil_linter.git
cd go_no_nil_linter

# Install dependencies
go mod download

# Generate protobuf code
cd proto && buf generate && cd ..
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -v ./analyzer -run TestAnalyzer
```

### Project Structure

- **`analyzer/`** - Core linter implementation
  - [`analyzer.go`](analyzer/analyzer.go:1) - Main analysis logic
  - [`messages.go`](analyzer/messages.go:1) - Protobuf type identification
  - [`detector.go`](analyzer/detector.go:1) - Nil detection and recursive validation
- **`cmd/nonillinter/`** - Command-line tool
- **`proto/`** - Protobuf definitions for testing
- **`gen/`** - Generated Go code from protobuf
- **`testdata/`** - Test cases (valid and invalid)

## Configuration

Currently, the linter uses sensible defaults and requires no configuration. Future versions may support:

- Custom message type patterns
- Configurable recursion depth
- Exclusion patterns
- Integration with golangci-lint

## Limitations

1. **Conservative nil detection** - May miss some complex dataflow patterns
2. **No cross-package analysis** - Only analyzes within a single package
3. **Struct tag parsing** - Relies on type system rather than parsing proto tags
4. **Optional field detection** - May need enhancement for complex optional patterns

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - See LICENSE file for details

## Related Projects

- [golang.org/x/tools/go/analysis](https://pkg.go.dev/golang.org/x/tools/go/analysis) - Go analysis framework
- [buf.build](https://buf.build/) - Modern protobuf tooling
- [golangci-lint](https://golangci-lint.run/) - Go linters aggregator

## Support

- **Issues**: https://github.com/nickheyer/go_no_nil_linter/issues
- **Discussions**: https://github.com/nickheyer/go_no_nil_linter/discussions

## Acknowledgments

Built with:
- [golang.org/x/tools/go/analysis](https://pkg.go.dev/golang.org/x/tools/go/analysis) - Static analysis framework
- [google.golang.org/protobuf](https://pkg.go.dev/google.golang.org/protobuf) - Protobuf runtime
- [buf.build](https://buf.build/) - Protobuf tooling