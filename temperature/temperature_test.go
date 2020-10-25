package temperature

import (
	"math/rand"
	"testing"
)

func TestGenerateTemperature(t *testing.T) {
	rand.Seed(42)
	b, err := GenerateTemperatureJSON(0, "test")
	if err != nil {
		t.Fatalf("Error from GenerateTemperatureJSON: %s", err)
	}
	temperature, err := ConvertTempToStruct(b)
	if err != nil {
		t.Fatalf("Error from ConvertTempToStruct: %s", err)
	}
	if temperature.Hostname != "test" {
		t.Fatalf("Expected: %s, Got %s", "test", temperature.Hostname)
	}
	if !(temperature.Celcius < 26 && temperature.Celcius > 25) {
		t.Fatalf("Expected: %s, Got: %v", "25.30..", temperature.Celcius)
	}
	t.Logf("temperature: %v\n", temperature)
}
