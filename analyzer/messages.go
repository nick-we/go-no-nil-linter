package analyzer

import (
	"go/types"
	"strings"
)

// isProtobufMessageType checks if a type is a protobuf message type
func isProtobufMessageType(t types.Type) bool {
	// Dereference pointer if needed
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	// Must be a named type
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}

	// Check if it has the ProtoMessage() method
	return hasProtoMessageMethod(named)
}

// hasProtoMessageMethod checks if a type has the ProtoMessage() method
func hasProtoMessageMethod(t *types.Named) bool {
	// Look for ProtoMessage() method
	for i := 0; i < t.NumMethods(); i++ {
		method := t.Method(i)
		if method.Name() == "ProtoMessage" {
			sig, ok := method.Type().(*types.Signature)
			if !ok {
				continue
			}
			// ProtoMessage() should have no params and no returns
			if sig.Params().Len() == 0 && sig.Results().Len() == 0 {
				return true
			}
		}
	}
	return false
}

// isMessageField checks if a struct field is a message type (not a scalar type)
func isMessageField(field *types.Var) bool {
	fieldType := field.Type()

	// Skip repeated fields (slices) - they are always optional
	if _, ok := fieldType.(*types.Slice); ok {
		return false
	}

	// Dereference pointer if needed
	if ptr, ok := fieldType.(*types.Pointer); ok {
		fieldType = ptr.Elem()
	}

	// Must be a named type
	named, ok := fieldType.(*types.Named)
	if !ok {
		return false
	}

	// Must have ProtoMessage() method
	if !hasProtoMessageMethod(named) {
		return false
	}

	// Get package path and type name
	obj := named.Obj()
	if obj == nil {
		return false
	}

	pkg := obj.Pkg()
	typeName := obj.Name()

	// Check if it's a well-known type (these are message types)
	if pkg != nil {
		pkgPath := pkg.Path()
		
		// Google protobuf well-known types
		if strings.Contains(pkgPath, "google.golang.org/protobuf/types/known") {
			return true
		}
		
		// Google API types (date, money, etc.)
		if strings.Contains(pkgPath, "google.golang.org/genproto/googleapis/type") {
			return true
		}
	}

	// Check if it's a scalar wrapper (these should be treated as scalars, not messages)
	scalarWrappers := map[string]bool{
		"StringValue":  true,
		"Int32Value":   true,
		"Int64Value":   true,
		"UInt32Value":  true,
		"UInt64Value":  true,
		"FloatValue":   true,
		"DoubleValue":  true,
		"BoolValue":    true,
		"BytesValue":   true,
	}

	if scalarWrappers[typeName] {
		return false // Scalar wrappers are optional by nature
	}

	// It's a custom message type
	return true
}

// isOptionalField checks if a field has the 'optional' keyword in proto3
func isOptionalField(field *types.Var) bool {
	fieldType := field.Type()
	
	// In proto3, optional message fields become **Type (double pointer)
	// Check if it's a pointer to a pointer
	if ptr, ok := fieldType.(*types.Pointer); ok {
		if _, ok := ptr.Elem().(*types.Pointer); ok {
			return true // Double pointer indicates optional in proto3
		}
		
		// Single pointer to a message type could be optional
		// In proto3, optional fields have specific characteristics
		// For a more robust check, we'd parse struct tags, but as a heuristic:
		// If it's a pointer to a message type, check if there's a corresponding Has method
		underlying := ptr.Elem()
		if named, ok := underlying.(*types.Named); ok {
			// Optional message fields in proto3 typically have Has<FieldName>() methods
			// This is a conservative check - if unsure, treat as required
			_ = named // Could check for Has methods here
		}
	}
	
	return false // Conservative: assume required unless we can prove optional
}

// getMessageFields returns all non-optional message fields from a struct type
func getMessageFields(structType *types.Struct) []*types.Var {
	var messageFields []*types.Var

	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)

		// Skip unexported fields
		if !field.Exported() {
			continue
		}

		// Check if it's a message field
		if !isMessageField(field) {
			continue
		}

		// Check if it's optional
		if isOptionalField(field) {
			continue
		}

		messageFields = append(messageFields, field)
	}

	return messageFields
}

// isWellKnownType checks if a type is a Google well-known type
func isWellKnownType(t types.Type) bool {
	// Dereference pointer if needed
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	named, ok := t.(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}

	pkgPath := obj.Pkg().Path()
	
	// Check for well-known types packages
	return strings.Contains(pkgPath, "google.golang.org/protobuf/types/known") ||
		strings.Contains(pkgPath, "google.golang.org/genproto/googleapis/type")
}