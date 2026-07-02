package service

type HealthStatus struct {
	Status string `json:"status"`
}

type HealthService struct{}

func NewHealthService() HealthService {
	return HealthService{}
}

func (HealthService) Status() HealthStatus {
	return HealthStatus{Status: "ok"}
}
