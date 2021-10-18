package akamai

type SensorRequest struct {
	SensorData string `json:"sensor_data"`
}

type SensorResponse struct {
	Success bool `json:"success"`
}
