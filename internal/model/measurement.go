package model

import "time"


type Measurement struct {
	SensorID  string    `json:"sensor_id"`
	Value     float64   `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}