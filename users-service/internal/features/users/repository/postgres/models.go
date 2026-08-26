package postgres

import "github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/domain"

type UserModel struct {
	ID          int
	Version     int
	FullName    string
	Email       string
	Password    string
	PhoneNumber *string
	IsAdmin     bool
	Status      string
}

func userDomainFromModels(users []UserModel) []domain.User {
	userDomain := make([]domain.User, len(users))
	for i, user := range users {
		userDomain[i] = domain.NewUser(user.ID, user.Version, user.FullName, user.Email, user.Password, user.PhoneNumber, user.IsAdmin, user.Status)
	}
	return userDomain
}
