package examples

import (
	examplev1 "github.com/nickheyer/go_no_nil_linter/gen/example/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This should trigger an error - user has uninitialized required fields
func badVariableUsage() *examplev1.UserResponse {
	// User created with missing required fields
	user := &examplev1.User{
		Id:   "123",
		Name: "John",
		// Address not initialized - VIOLATION
		// CreatedAt not initialized - VIOLATION  
		// ContactInfo not initialized - VIOLATION
	}

	// Using this user in a response should be flagged
	response := &examplev1.UserResponse{
		User:      user, // Should detect user has uninitialized fields
		LastLogin: timestamppb.Now(),
	}
	return response
}

// This should also trigger an error - nil variable
func nilVariableUsage() *examplev1.UserResponse {
	var user *examplev1.User // user is nil
	
	response := &examplev1.UserResponse{
		User:      user, // Should detect user is nil
		LastLogin: timestamppb.Now(),
	}
	return response
}

// This should be OK - user is properly initialized
func goodVariableUsage() *examplev1.UserResponse {
	user := &examplev1.User{
		Id:   "123",
		Name: "John",
		Address: &examplev1.Address{
			Street:     "123 Main St",
			City:       "NYC",
			PostalCode: "10001",
			Location: &examplev1.Location{
				Latitude:  40.7128,
				Longitude: -74.0060,
			},
		},
		CreatedAt: timestamppb.Now(),
		ContactInfo: &examplev1.ContactInfo{
			Email: "john@example.com",
			Phone: "555-1234",
		},
	}

	response := &examplev1.UserResponse{
		User:      user, // OK - user is fully initialized
		LastLogin: timestamppb.Now(),
	}
	return response
}