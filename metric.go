package co2_monitor

import (
	"fmt"
	"time"
)

type MetricType string

const (
	Unknown MetricType = "UNKNOWN"
	CO2     MetricType = "CO2"
	Temp    MetricType = "TEMPERATURE"
	Error   MetricType = "ERROR"
)

type Metric struct {
	Type  MetricType
	Value string
	Time  time.Time
}

func (m Metric) String() string {
	return fmt.Sprintf("[%v] %s: %v", m.Time, m.Type, m.Value)
}
