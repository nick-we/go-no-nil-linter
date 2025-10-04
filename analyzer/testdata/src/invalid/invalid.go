package invalid

import (
	examplev1 "github.com/nickheyer/go_no_nil_linter/gen/example/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func explicitNilAssignment() {
	response := &examplev1.UserResponse{}
	response.User = nil // want "nil assignment to non-optional message field 'User'"
}

func implicitNilAssignment() {
	var user *examplev1.User
	response := &examplev1.UserResponse{}
	response.User = user // want "variable 'user' used for field 'User' is nil"
}

func nilInCompositeLiteral() {
	_ = &examplev1.UserResponse{
		User:      nil, // want "nil assignment to non-optional message field 'User'"
		LastLogin: nil, // want "nil assignment to non-optional message field 'LastLogin'"
	}
}

func uninitializedMessageField() {
	_ = &examplev1.UserResponse{ // want "non-optional message field 'User' not initialized" "non-optional message field 'LastLogin' not initialized"
		RelatedUsers: []*examplev1.User{},
	}
}

func nestedNilAssignment() {
	user := &examplev1.User{
		Id:   "123",
		Name: "John",
		Address: nil, // want "nil assignment to non-optional message field 'Address'"
	}
	_ = &examplev1.UserResponse{
		User:      user,
		LastLogin: nil, // want "nil assignment to non-optional message field 'LastLogin'"
	}
}

func deeplyNestedNil() {
	addr := &examplev1.Address{
		Street:     "123 Main St",
		City:       "NYC",
		PostalCode: "10001",
		Location:   nil, // want "nil assignment to non-optional message field 'Location'"
	}

	user := &examplev1.User{
		Id:      "123",
		Name:    "John",
		Address: addr, // This triggers recursive check
	}

	_ = &examplev1.UserResponse{
		User:      user, // want "nil assignment to non-optional message field 'User.Address.Location'"
		LastLogin: nil,  // want "nil assignment to non-optional message field 'LastLogin'"
	}
}

func missingContactInfo() {
	user := &examplev1.User{ // want "non-optional message field 'Address' not initialized" "non-optional message field 'CreatedAt' not initialized" "non-optional message field 'ContactInfo' not initialized"
		Id:   "123",
		Name: "John",
	}
	_ = &examplev1.UserResponse{
		User:      user,
		LastLogin: nil, // want "nil assignment to non-optional message field 'LastLogin'"
	}
}

func nilWellKnownType() {
	user := &examplev1.User{
		Id:        "123",
		Name:      "John",
		CreatedAt: nil, // want "nil assignment to non-optional message field 'CreatedAt'"
	}
	response := &examplev1.UserResponse{}
	response.User = user
}

func assignmentAfterCreation() {
	response := &examplev1.UserResponse{
		User: &examplev1.User{
			Id:   "123",
			Name: "John",
		},
		LastLogin: nil, // want "nil assignment to non-optional message field 'LastLogin'"
	}

	// Later assignment
	response.User.Address = nil // want "nil assignment to non-optional message field 'Address'"
}

func nilInNestedStruct() {
	contact := &examplev1.ContactInfo{
		Email: "test@example.com",
		Phone: "555-1234",
		// MailingAddress is optional, so nil is OK
	}

	user := &examplev1.User{
		Id:          "123",
		Name:        "John",
		ContactInfo: contact,
		CreatedAt:   timestamppb.Now(),
	}

	_ = &examplev1.UserResponse{
		User:      user,        // want "non-optional message field 'User.Address' not initialized"
		LastLogin: nil,         // want "nil assignment to non-optional message field 'LastLogin'"
	}
}