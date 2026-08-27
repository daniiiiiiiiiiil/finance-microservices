package ports

type JWTManagerInterface interface {
	Generate(userID int, email string, isAdmin bool) (string, error)
}
