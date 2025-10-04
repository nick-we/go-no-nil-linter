package valid

import (
	examplev1 "github.com/nickheyer/go_no_nil_linter/gen/example/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func fullyInitializedResponse() {
	response := &examplev1.UserResponse{
		User: &examplev1.User{
			Id:   "123",
			Name: "John Doe",
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
				Email: "john@example.com",
				Phone: "555-1234",
			},
		},
		LastLogin:    timestamppb.Now(),
		RelatedUsers: []*examplev1.User{},
	}
	_ = response
}

func optionalFieldCanBeNil() {
	user := &examplev1.User{
		Id:       "123",
		Name:     "John",
		Nickname: nil, // Optional field - OK to be nil
		Address: &examplev1.Address{
			Street:     "123 Main St",
			City:       "NYC",
			PostalCode: "10001",
			Location: &examplev1.Location{
				Latitude:  40.7128,
				Longitude: -74.0060,
			},
			Apartment: nil, // Optional field - OK to be nil
		},
		CreatedAt: timestamppb.Now(),
		ContactInfo: &examplev1.ContactInfo{
			Email:          "john@example.com",
			Phone:          "555-1234",
			MailingAddress: nil, // Optional message field - OK to be nil
		},
	}

	response := &examplev1.UserResponse{
		User:      user,
		LastLogin: timestamppb.Now(),
		Manager:   nil, // Optional message field - OK to be nil
	}
	_ = response
}

func scalarFieldsCanBeZero() {
	// Scalar fields can have zero values - not checked by linter
	user := &examplev1.User{
		Id:   "", // Empty string is OK for scalars
		Name: "", // Empty string is OK for scalars
		Address: &examplev1.Address{
			Street:     "",
			City:       "",
			PostalCode: "",
			Location: &examplev1.Location{
				Latitude:  0.0,
				Longitude: 0.0,
			},
		},
		CreatedAt: timestamppb.Now(),
		ContactInfo: &examplev1.ContactInfo{
			Email: "",
			Phone: "",
		},
	}

	response := &examplev1.UserResponse{
		User:      user,
		LastLogin: timestamppb.Now(),
	}
	_ = response
}

func assignmentFromFunction() examplev1.UserResponse {
	return examplev1.UserResponse{
		User:      createUser(),
		LastLogin: timestamppb.Now(),
	}
}

func createUser() *examplev1.User {
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
}

func updateExistingResponse() {
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

	// Update with new valid values
	response.User.Name = "Jane"
	response.User.Address.City = "Boston"
	response.User.Address.Location = &examplev1.Location{
		Latitude:  42.3601,
		Longitude: -71.0589,
	}

	_ = response
}

func repeatedFieldsWithMessages() {
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
		RelatedUsers: []*examplev1.User{
			{
				Id:   "456",
				Name: "Jane",
				Address: &examplev1.Address{
					Street:     "456 Oak Ave",
					City:       "Boston",
					PostalCode: "02101",
					Location: &examplev1.Location{
						Latitude:  42.3601,
						Longitude: -71.0589,
					},
				},
				CreatedAt: timestamppb.Now(),
				ContactInfo: &examplev1.ContactInfo{
					Email: "jane@example.com",
					Phone: "555-5678",
				},
			},
		},
	}
	_ = response
}

func listResponse() {
	response := &examplev1.ListUsersResponse{
		Users: []*examplev1.User{
			{
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
		},
		FetchedAt: timestamppb.Now(),
	}
	_ = response
}