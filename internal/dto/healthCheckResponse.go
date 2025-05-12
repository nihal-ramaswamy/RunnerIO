package dto

type HealthCheckResponse struct {
	Message string
}

func NewHealthCheckResponse(message string) HealthCheckResponse {
	return HealthCheckResponse{
		Message: message,
	}
}
