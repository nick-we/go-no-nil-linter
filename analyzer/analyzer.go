package analyzer

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer is the main analyzer for detecting nil assignments to non-optional protobuf message fields
var Analyzer = &analysis.Analyzer{
	Name:     "nonillinter",
	Doc:      "detects nil assignments to non-optional protobuf message fields",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Skip generated protobuf files (.pb.go)
	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename
		if strings.HasSuffix(filename, ".pb.go") {
			return nil, nil
		}
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Track analyzed composite literals to avoid duplicate checks
	analyzedComposites := make(map[ast.Node]bool)

	// Node types we care about
	nodeFilter := []ast.Node{
		(*ast.AssignStmt)(nil),   // Regular assignments
		(*ast.CompositeLit)(nil), // Struct literals
		(*ast.ReturnStmt)(nil),   // Return statements
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			checkAssignment(stmt, pass)

		case *ast.CompositeLit:
			// Avoid duplicate analysis if we've already checked this composite
			if analyzedComposites[stmt] {
				return
			}
			analyzedComposites[stmt] = true

			// Check if this is creating a protobuf message type
			litType := pass.TypesInfo.TypeOf(stmt)
			if litType == nil {
				return
			}

			if isResponseMessage(litType) {
				checkCompositeLiteral(stmt, litType, pass)
			}

		case *ast.ReturnStmt:
			// Check return statements for composite literals creating messages
			for _, result := range stmt.Results {
				if comp, ok := result.(*ast.CompositeLit); ok {
					if analyzedComposites[comp] {
						continue
					}
					analyzedComposites[comp] = true

					litType := pass.TypesInfo.TypeOf(comp)
					if litType != nil && isResponseMessage(litType) {
						checkCompositeLiteral(comp, litType, pass)
					}
				}
			}
		}
	})

	return nil, nil
}

// checkAssignment checks an assignment statement for nil assignments to message fields
func checkAssignment(stmt *ast.AssignStmt, pass *analysis.Pass) {
	for i := 0; i < len(stmt.Lhs) && i < len(stmt.Rhs); i++ {
		lhs := stmt.Lhs[i]
		rhs := stmt.Rhs[i]

		// Check if LHS is a selector expression (field access)
		sel, ok := lhs.(*ast.SelectorExpr)
		if !ok {
			continue
		}

		// Get the type of the base expression
		baseType := pass.TypesInfo.TypeOf(sel.X)
		if baseType == nil {
			continue
		}

		// Dereference pointer types
		if ptr, ok := baseType.(*types.Pointer); ok {
			baseType = ptr.Elem()
		}

		// Check if the base is a response message type - only check response messages
		if !isResponseMessage(baseType) {
			continue
		}

		// Get the field being accessed
		field := getFieldFromType(baseType, sel.Sel.Name)
		if field == nil {
			continue
		}

		// Check if this is a message field (not scalar)
		if !isMessageField(field) {
			continue
		}

		// Check if the field is optional
		if isOptionalField(field) {
			continue
		}

		// Check if RHS is nil (explicit or implicit)
		if isNilValue(rhs, pass) {
			pass.Reportf(rhs.Pos(),
				"nil assignment to non-optional message field '%s' in protobuf message '%s'",
				sel.Sel.Name, baseType.String())
		} else {
			// If RHS is not nil but is a message type, recursively validate it
			rhsType := pass.TypesInfo.TypeOf(rhs)
			if rhsType != nil && isProtobufMessageType(rhsType) {
				validateMessageValue(rhs, rhsType, pass, sel.Sel.Name)
			}
		}
	}
}

// checkCompositeLiteral checks a composite literal for nil message fields
func checkCompositeLiteral(lit *ast.CompositeLit, litType types.Type, pass *analysis.Pass) {
	// Only check if this is a response message type
	if !isResponseMessage(litType) {
		return
	}

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
			// Handle positional initialization if needed
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
				"nil assignment to non-optional message field '%s' in protobuf message '%s'",
				fieldName, litType.String())
		} else {
			// Recursively validate non-nil message values
			valueType := pass.TypesInfo.TypeOf(kv.Value)
			if valueType != nil && isProtobufMessageType(valueType) {
				validateMessageValue(kv.Value, valueType, pass, fieldName)
			}
		}
	}

	// Check for uninitialized required message fields
	for _, field := range messageFields {
		if !initialized[field.Name()] {
			pass.Reportf(lit.Pos(),
				"non-optional message field '%s' not initialized in protobuf message '%s'",
				field.Name(), litType.String())
		}
	}
}

// getStructType extracts the struct type from a type, handling pointers
func getStructType(t types.Type) *types.Struct {
	// Dereference pointer if needed
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	// Get named type
	named, ok := t.(*types.Named)
	if !ok {
		return nil
	}

	// Get underlying struct
	structType, ok := named.Underlying().(*types.Struct)
	if !ok {
		return nil
	}

	return structType
}

// getFieldFromType gets a field by name from a type
func getFieldFromType(t types.Type, fieldName string) *types.Var {
	structType := getStructType(t)
	if structType == nil {
		return nil
	}

	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		if field.Name() == fieldName {
			return field
		}
	}

	return nil
}