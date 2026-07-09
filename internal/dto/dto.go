package dto

type UserRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,max=72"`
}

type UserResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ResponseToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
