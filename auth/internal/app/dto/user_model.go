package dto

type RegisterUserCommand struct {
	Email    string
	Password string
	Role     string
}

type LoginCommand struct {
	Email       string
	Password    string
	PhoneNumber string
}
