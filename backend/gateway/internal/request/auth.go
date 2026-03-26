package request

type LoginRequest struct {
	Phone     string `json:"phone" binding:"required,phoneValidator"`
	LoginType int    `json:"login_type" binding:"required"` // 1密码 2验证码
	Password  string `json:"password"`                      // login_type=1时必填
	Code      string `json:"code"`                          // login_type=2时必填
}

type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required,phoneValidator"`
	Password string `json:"password"  binding:"required"`
	Code     string `json:"code"  binding:"required"`
}
