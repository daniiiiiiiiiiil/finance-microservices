package users

type CreateProfileRequest struct {
	Email       string
	FullName    string
	PhoneNumber *string
	IsAdmin     bool
}

type UserProfile struct {
	ID          int
	FullName    string
	Email       string
	PhoneNumber *string
	IsAdmin     bool
	IsActive    bool
}

type ListUsersResponse struct {
	Users  []UserProfile
	Limit  int
	Offset int
}

type UserMetrics struct {
	TotalUsers int
}
