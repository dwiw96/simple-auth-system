package delivery

import auth "github.com/dwiw96/simple-auth-system/features/auth"

type signupResponse struct {
	FullName      string `json:"fullname"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	Gender        string `json:"gender"`
	MaritalStatus string `json:"marital_status"`
}

func toSignUpResponse(input *auth.User) signupResponse {
	return signupResponse{
		FullName:      input.Fullname,
		Email:         input.Email,
		Address:       input.Address,
		Gender:        input.Gender,
		MaritalStatus: input.MaritalStatus,
	}
}
