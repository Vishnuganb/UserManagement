package model

type User struct {
	ID        int64   `json:"id"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Email     string  `json:"email"`
	Phone     *string `json:"phone,omitempty"`
	Age       *int32  `json:"age,omitempty"`
	Status    *string `json:"status,omitempty"`
}

type CreateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Age       int    `json:"age"`
	Status    string `json:"status"`
}

type UpdateUserRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
	Age       *int32  `json:"age"`
	Status    *string `json:"status"`
}

type CUDRequest struct {
	Type      string
	CreateReq CreateUserRequest
	UpdateReq struct {
		UserID int64
		Req    UpdateUserRequest
	}
	UserID          int64
	ResponseChannel chan interface{}
}
