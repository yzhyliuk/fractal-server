package apimodels

type ResetPassword struct {
	NewPassword string `json:"newPassword"`
	ResetCode string `json:"resetCode"`
	UserID int `json:"userId"`
}
