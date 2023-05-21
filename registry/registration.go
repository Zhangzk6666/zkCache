package registry

type GetServiceDTO struct {
	Url string `form:"url" json:"url" validate:"required"`
}

type RegistrationVO struct {
	ServiceName ServiceName `form:"serviceName" json:"serviceName" validate:"required"`
	ServiceURL  string      `form:"serviceURL" json:"serviceURL" validate:"required"`
}

type GetServiceVO struct {
	ServiceName ServiceName `form:"serviceName" json:"serviceName" validate:"required"`
	Key         string      `form:"key" json:"key" validate:"required"`
}

// type Registration struct {
// 	ServiceName ServiceName
// 	ServiceURL  string
// }

type ServiceName string

const (
	LogService = ServiceName("LogService")
)
