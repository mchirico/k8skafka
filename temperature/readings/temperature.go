package readings

import (
	"encoding/json"
	"math/rand"
	"time"
)

// celcius-readings
type Temperature struct {
	ID        int64
	Hostname  string
	Celcius   float64
	TimeStamp int64
}

func GenerateTemperatureJSON(id int64, hostname string) ([]byte, error) {
	timeStamp := time.Now().Unix()
	celcius := rand.Float64()*100 - 12
	reading := Temperature{id, hostname, celcius, timeStamp}
	return json.Marshal(reading)

}

func ConvertTempToStruct(b []byte) (Temperature, error) {
	var temperature Temperature
	err := json.Unmarshal(b, &temperature)
	return temperature, err
}
