package utils

// === AUTH ===
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CategoriesRequest struct {
	Name string `json:"name"`
}

type RequestTransactions struct {
	Phone            string `json:"phone" example:"081234567890"`
	Email            string `json:"email" example:"youremail@gmail.com"`
	Address          string `json:"address" example:"jakarta"`
	Fullname         string `json:"fullname" example:"fullname"`
	Id_PaymentMethod int    `json:"id_paymentMethod"`
	Id_Delivery      int    `json:"id_delivery"`
}
