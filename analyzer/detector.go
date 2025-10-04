package analyzer

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// isNilValue checks if an expression evaluates to nil
func isNilValue(expr ast.Expr, pass *analysis.Pass) bool {
	// Check for nil literal
	if ident, ok := expr.(*ast.Ident); ok && ident.Name == "nil" {
		return true
	}

	// Check for typed nil: (*Type)(nil)
	if call, ok := expr.(*ast.CallExpr); ok {
		if len(call.Args) == 1 {
			if ident, ok := call.Args[0].(*ast.Ident); ok && ident.Name == "nil" {
				return true
			}
		}
	}

	// Check for variable that might be nil
	if ident, ok := expr.(*ast.Ident); ok {
		if isNilVariable(ident, pass) {
			return true
		}
	}

	// Check for unary expression: &nil (although this is invalid Go)
	if unary, ok := expr.(*ast.UnaryExpr); ok {
		if unary.Op == token.AND {
			return isNilValue(unary.X, pass)
		}
	}

	return false
}

// isNilVariable checks if a variable identifier is nil
func isNilVariable(ident *ast.Ident, pass *analysis.Pass) bool {
	// Get the object this identifier refers to
	obj := pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return false
	}

	// Try to get the constant value
	if tv, ok := pass.TypesInfo.Types[ident]; ok {
		if tv.Value != nil {
			return false // Has a constant value, not nil
		}
	}

	// Try to find the variable declaration
	var decl *ast.ValueSpec
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			if vs, ok := n.(*ast.ValueSpec); ok {
				for _, name := range vs.Names {
					if pass.TypesInfo.ObjectOf(name) == obj {
						decl = vs
						return false
					}
				}
			}
			return true
		})
		if decl != nil {
			break
		}
	}

	if decl == nil {
		// Could be a parameter or return value, assume not nil
		return false
	}

	// Check if it has an initializer
	if len(decl.Values) == 0 {
		// No initializer means zero value
		// For pointers and interfaces, zero value is nil
		objType := obj.Type()
		if _, ok := objType.(*types.Pointer); ok {
			return true
		}
		if _, ok := objType.(*types.Interface); ok {
			return true
		}
		return false
	}

	// Check if the initializer is nil
	for _, value := range decl.Values {
		if isNilValue(value, pass) {
			return true
		}
	}

	return false
}

// validateMessageValue recursively validates a message value for nil fields
func validateMessageValue(expr ast.Expr, exprType types.Type, pass *analysis.Pass, fieldContext string) {
	switch e := expr.(type) {
	case *ast.Ident:
		// Variable reference - try to trace to its declaration
		validateVariableMessage(e, exprType, pass, fieldContext)

	case *ast.CompositeLit:
		// Struct literal - check its fields recursively
		validateCompositeLiteralMessage(e, exprType, pass, fieldContext)

	case *ast.CallExpr:
		// Function call - we can't easily analyze what it returns
		// Conservative approach: assume it's valid
		return

	case *ast.UnaryExpr:
		// Address operation (&expr)
		if e.Op == token.AND {
			validateMessageValue(e.X, exprType, pass, fieldContext)
		}

	case *ast.SelectorExpr:
		// Field access - validate the selected field
		// This is already a reference to an existing object, assume valid
		return
	}
}

// validateVariableMessage traces a variable to its declaration and validates it
func validateVariableMessage(ident *ast.Ident, exprType types.Type, pass *analysis.Pass, fieldContext string) {
	obj := pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return
	}

	// Find the variable declaration
	var decl *ast.ValueSpec
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			if vs, ok := n.(*ast.ValueSpec); ok {
				for _, name := range vs.Names {
					if pass.TypesInfo.ObjectOf(name) == obj {
						decl = vs
						return false
					}
				}
			}
			return true
		})
		if decl != nil {
			break
		}
	}

	if decl == nil {
		// Could be a parameter, assume valid
		return
	}

	// If no initializer, it's zero value (nil for pointers)
	if len(decl.Values) == 0 {
		if _, ok := exprType.(*types.Pointer); ok {
			pass.Reportf(ident.Pos(),
				"variable '%s' used for field '%s' is nil (zero value)",
				ident.Name, fieldContext)
		}
		return
	}

	// Recursively validate the initializer
	for i, name := range decl.Names {
		if pass.TypesInfo.ObjectOf(name) == obj && i < len(decl.Values) {
			value := decl.Values[i]
			if comp, ok := value.(*ast.CompositeLit); ok {
				validateCompositeLiteralMessage(comp, exprType, pass, fieldContext)
			}
		}
	}
}

// validateCompositeLiteralMessage recursively validates a composite literal
func validateCompositeLiteralMessage(lit *ast.CompositeLit, litType types.Type, pass *analysis.Pass, fieldContext string) {
	// Get the struct type
	structType := getStructType(litType)
	if structType == nil {
		return
	}

	// Get all message fields for this type
	messageFields := getMessageFields(structType)
	if len(messageFields) == 0 {
		return
	}

	// Track which fields are initialized
	initialized := make(map[string]bool)

	// Check each element in the composite literal
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		// Get the field name
		fieldIdent, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}

		fieldName := fieldIdent.Name
		initialized[fieldName] = true

		// Find the corresponding field
		var field *types.Var
		for _, f := range messageFields {
			if f.Name() == fieldName {
				field = f
				break
			}
		}

		if field == nil {
			continue
		}

		// Check if value is nil
		if isNilValue(kv.Value, pass) {
			pass.Reportf(kv.Value.Pos(),
				"nil assignment to non-optional message field '%s.%s' in protobuf message '%s'",
				fieldContext, fieldName, litType.String())
		} else {
			// Recursively validate non-nil message values
			valueType := pass.TypesInfo.TypeOf(kv.Value)
			if valueType != nil && isProtobufMessageType(valueType) {
				nestedContext := fieldContext + "." + fieldName
				validateMessageValue(kv.Value, valueType, pass, nestedContext)
			}
		}
	}

	// Check for uninitialized required message fields
	for _, field := range messageFields {
		if !initialized[field.Name()] {
			pass.Reportf(lit.Pos(),
				"non-optional message field '%s.%s' not initialized in protobuf message '%s'",
				fieldContext, field.Name(), litType.String())
		}
	}
}