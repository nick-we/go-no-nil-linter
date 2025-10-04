package analyzer

import (
	"go/types"
	"strings"
)

// isResponseMessage checks if a type is a protobuf response message
// Response messages are types that are returned from service endpoints
func isResponseMessage(t types.Type) bool {
	// Dereference pointer if needed
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	// Must be a named type
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}

	// Must be a protobuf message type
	if !hasProtoMessageMethod(named) {
		return false
	}

	// Get the type name
	obj := named.Obj()
	if obj == nil {
		return false
	}

	typeName := obj.Name()

	// Check if it matches response naming convention
	// Response messages typically end with "Response"
	if strings.HasSuffix(typeName, "Response") {
		return true
	}

	// Could also check for other patterns like "*Reply", "*Result", etc.
	if strings.HasSuffix(typeName, "Reply") {
		return true
	}

	if strings.HasSuffix(typeName, "Result") {
		return true
	}

	return false
}

// shouldCheckType determines if we should check this type for nil fields
// We only check response messages and their submessages
func shouldCheckType(t types.Type) bool {
	return isResponseMessage(t)
}