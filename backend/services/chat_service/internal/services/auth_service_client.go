package services

type AuthClient struct {
	BaseURL string
}

func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{
		BaseURL: baseURL,
	}
}

// Kullanıcının geçerli olup olmadığını kontrol eder
// Geçici olarak tüm kullanıcıları geçerli sayıyoruz
func (a *AuthClient) IsUserValid(userID string) bool {
	return true
}
