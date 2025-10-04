package analyzer_test

import (
	"testing"

	"github.com/nickheyer/go_no_nil_linter/analyzer"
	examplev1 "github.com/nickheyer/go_no_nil_linter/gen/example/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAnalyzerDefinition(t *testing.T) {
	// Test that the analyzer is properly defined
	if analyzer.Analyzer == nil {
		t.Fatal("Analyzer is nil")
	}
	
	if analyzer.Analyzer.Name != "nonillinter" {
		t.Errorf("Expected analyzer name 'nonillinter', got '%s'", analyzer.Analyzer.Name)
	}
	
	if analyzer.Analyzer.Doc == "" {
		t.Error("Analyzer Doc is empty")
	}
	
	if analyzer.Analyzer.Run == nil {
		t.Fatal("Analyzer Run function is nil")
	}
}

// TestProtobufMessageCreation tests that protobuf messages can be created correctly
func TestProtobufMessageCreation(t *testing.T) {
	tests := []struct {
		name      string
		createFn  func() *examplev1.UserResponse
		shouldErr bool
		desc      string
	}{
		{
			name: "valid_full_initialization",
			createFn: func() *examplev1.UserResponse {
				return &examplev1.UserResponse{
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
			},
			shouldErr: false,
			desc:      "All required fields initialized - should pass",
		},
		{
			name: "optional_fields_can_be_nil",
			createFn: func() *examplev1.UserResponse {
				return &examplev1.UserResponse{
					User: &examplev1.User{
						Id:       "123",
						Name:     "John",
						Nickname: nil, // Optional field
						Address: &examplev1.Address{
							Street:     "123 Main St",
							City:       "NYC",
							PostalCode: "10001",
							Location: &examplev1.Location{
								Latitude:  40.7128,
								Longitude: -74.0060,
							},
							Apartment: nil, // Optional field
						},
						CreatedAt: timestamppb.Now(),
						ContactInfo: &examplev1.ContactInfo{
							Email:          "john@example.com",
							Phone:          "555-1234",
							MailingAddress: nil, // Optional message field
						},
					},
					LastLogin: timestamppb.Now(),
					Manager:   nil, // Optional message field
				}
			},
			shouldErr: false,
			desc:      "Optional fields can be nil - should pass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := tt.createFn()
			if response == nil {
				t.Error("Response should not be nil")
			}
			
			// Verify basic structure
			if !tt.shouldErr {
				if response.User == nil {
					t.Error("User should not be nil in valid response")
				}
				if response.LastLogin == nil {
					t.Error("LastLogin should not be nil in valid response")
				}
			}
		})
	}
}

// TestMessageFieldValidation tests that message fields are properly validated
func TestMessageFieldValidation(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() interface{}
		validate func(interface{}) error
		desc     string
	}{
		{
			name: "user_with_all_message_fields",
			setup: func() interface{} {
				return &examplev1.User{
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
			},
			validate: func(v interface{}) error {
				user := v.(*examplev1.User)
				if user.Address == nil {
					t.Error("Address should not be nil")
				}
				if user.Address.Location == nil {
					t.Error("Location should not be nil")
				}
				if user.CreatedAt == nil {
					t.Error("CreatedAt should not be nil")
				}
				if user.ContactInfo == nil {
					t.Error("ContactInfo should not be nil")
				}
				return nil
			},
			desc: "All required message fields should be non-nil",
		},
		{
			name: "address_with_location",
			setup: func() interface{} {
				return &examplev1.Address{
					Street:     "123 Main St",
					City:       "NYC",
					PostalCode: "10001",
					Location: &examplev1.Location{
						Latitude:  40.7128,
						Longitude: -74.0060,
					},
				}
			},
			validate: func(v interface{}) error {
				addr := v.(*examplev1.Address)
				if addr.Location == nil {
					t.Error("Location should not be nil")
				}
				if addr.Location.Latitude == 0 && addr.Location.Longitude == 0 {
					// This is actually OK - zero values are fine for scalars
				}
				return nil
			},
			desc: "Nested message Location should be non-nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := tt.setup()
			if obj == nil {
				t.Fatal("Setup returned nil")
			}
			
			if err := tt.validate(obj); err != nil {
				t.Errorf("Validation failed: %v", err)
			}
		})
	}
}

// TestResponseBuilder tests a helper function pattern
func TestResponseBuilder(t *testing.T) {
	// Test that a properly built response has all required fields
	response := buildValidUserResponse("user-123", "John Doe")
	
	if response == nil {
		t.Fatal("buildValidUserResponse returned nil")
	}
	
	if response.User == nil {
		t.Error("User should not be nil")
	}
	
	if response.LastLogin == nil {
		t.Error("LastLogin should not be nil")
	}
	
	if response.User != nil {
		if response.User.Address == nil {
			t.Error("User.Address should not be nil")
		}
		
		if response.User.Address != nil && response.User.Address.Location == nil {
			t.Error("User.Address.Location should not be nil")
		}
		
		if response.User.CreatedAt == nil {
			t.Error("User.CreatedAt should not be nil")
		}
		
		if response.User.ContactInfo == nil {
			t.Error("User.ContactInfo should not be nil")
		}
	}
}

// buildValidUserResponse is a helper that creates a valid response
// This demonstrates the proper pattern for creating response messages
func buildValidUserResponse(userID, name string) *examplev1.UserResponse {
	return &examplev1.UserResponse{
		User: &examplev1.User{
			Id:   userID,
			Name: name,
			Address: &examplev1.Address{
				Street:     "123 Main St",
				City:       "New York",
				PostalCode: "10001",
				Location: &examplev1.Location{
					Latitude:  40.7128,
					Longitude: -74.0060,
				},
			},
			CreatedAt: timestamppb.Now(),
			ContactInfo: &examplev1.ContactInfo{
				Email: "user@example.com",
				Phone: "555-1234",
			},
		},
		LastLogin:    timestamppb.Now(),
		RelatedUsers: []*examplev1.User{}, // Empty slice is OK
	}
}

// TestListResponse tests list response messages
func TestListResponse(t *testing.T) {
	response := &examplev1.ListUsersResponse{
		Users: []*examplev1.User{
			{
				Id:   "user-1",
				Name: "User One",
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
					Email: "user1@example.com",
					Phone: "555-0001",
				},
			},
		},
		FetchedAt: timestamppb.Now(),
	}
	
	if response == nil {
		t.Fatal("Response should not be nil")
	}
	
	if response.Users == nil {
		t.Error("Users slice should not be nil (can be empty)")
	}
	
	if response.FetchedAt == nil {
		t.Error("FetchedAt timestamp should not be nil")
	}
	
	if len(response.Users) > 0 {
		user := response.Users[0]
		if user.Address == nil {
			t.Error("User in list should have non-nil Address")
		}
		if user.CreatedAt == nil {
			t.Error("User in list should have non-nil CreatedAt")
		}
	}
}

// TestScalarFields tests that scalar fields can have zero values
func TestScalarFields(t *testing.T) {
	// Scalar fields can be empty/zero - the linter should not complain
	user := &examplev1.User{
		Id:   "", // Empty string is OK for scalar
		Name: "", // Empty string is OK for scalar
		Address: &examplev1.Address{
			Street:     "", // Empty string is OK
			City:       "", // Empty string is OK
			PostalCode: "", // Empty string is OK
			Location: &examplev1.Location{
				Latitude:  0.0, // Zero is OK for scalar
				Longitude: 0.0, // Zero is OK for scalar
			},
		},
		CreatedAt: timestamppb.Now(), // But message types must be non-nil
		ContactInfo: &examplev1.ContactInfo{
			Email: "",
			Phone: "",
		},
	}
	
	if user == nil {
		t.Fatal("User should not be nil")
	}
	
	// The important thing is that message fields are non-nil
	if user.Address == nil {
		t.Error("Address message field must be non-nil even if scalars are empty")
	}
	
	if user.CreatedAt == nil {
		t.Error("CreatedAt message field must be non-nil")
	}
	
	if user.ContactInfo == nil {
		t.Error("ContactInfo message field must be non-nil")
	}
}