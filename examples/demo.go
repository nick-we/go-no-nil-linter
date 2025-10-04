package examples

import (
	examplev1 "github.com/nickheyer/go_no_nil_linter/gen/example/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This function works with a non-response message - NOT CHECKED by linter
func createUser() *examplev1.User {
	user := &examplev1.User{
		Id:   "123",
		Name: "John",
		// Address: nil,  // This would NOT be flagged - User is not a Response message
	}
	return user
}

// This function works with a response message - CHECKED by linter
func createUserResponse() *examplev1.UserResponse {
	response := &examplev1.UserResponse{
		User: &examplev1.User{
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
		},
		LastLogin: timestamppb.Now(),
	}
	return response
}

// This would trigger a violation because it's a Response message
func createBadResponse() *examplev1.UserResponse {
	// This WILL be flagged by the linter
	return &examplev1.UserResponse{
		// User field not initialized - VIOLATION
		LastLogin: timestamppb.Now(),
	}
}