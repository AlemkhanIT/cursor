package services

// MockEmailService is a mock implementation of EmailService for testing
type MockEmailService struct{}

func NewMockEmailService() EmailServiceInterface {
	return &MockEmailService{}
}

func (s *MockEmailService) GenerateToken() (string, error) {
	return "mock-token-123456789", nil
}

func (s *MockEmailService) SendEmailConfirmation(email, token string) error {
	// Mock implementation - just return nil (success)
	return nil
}

func (s *MockEmailService) SendPasswordReset(email, token string) error {
	// Mock implementation - just return nil (success)
	return nil
}
