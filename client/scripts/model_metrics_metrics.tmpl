// Modified after generation
package models

import (
	"encoding/json"
)

// MetricsMetrics struct for MetricsMetrics
type MetricsMetrics struct {
	// Metric details.
	Name         string      `json:"name,omitempty"`
	Values       interface{} `json:"values,omitempty"`
	MetricValues []MetricsValues
}

// UnmarshalJSON - unmarshal json for MetricsMetrics
func (m *MetricsMetrics) UnmarshalJSON(bytes []byte) error {
	d := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &d); err != nil {
		return err
	}

	m.Name = d["name"].(string)
	m.Values = d["values"].([]interface{})

	valBytes, err := json.Marshal(m.Values)
	if err != nil {
		return err
	}

	v := make([]MetricsValues, 0)
	if err := json.Unmarshal(valBytes, &v); err != nil {
		return nil
	}
	m.MetricValues = v

	return nil
}
