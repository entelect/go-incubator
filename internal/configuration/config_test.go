package config

import (
	"os"
	"testing"
)

func TestReadConfig(t *testing.T) {
	os.Setenv("TEST_HTTPPORT", "1234")
	os.Setenv("INVALID_HTTPPORT", "abcd")

	// Test that missing config results in an error
	_, err := ReadConfig("MISSING_")
	if err == nil {
		t.Error("missing config not returning an error")
	}

	// Test that invalid config results in an error
	_, err = ReadConfig("INVALID_")
	if err == nil {
		t.Error("invalid config not returning an error")
	}

	// Test that valid config returns as expected
	want := Configuration{HttpPort: 1234}
	got, err := ReadConfig("TEST_")
	if err != nil {
		t.Errorf("valid config returning an eror: %v", err)
	}
	if got != want {
		t.Errorf("config %+v not as expected %+v", got, want)
	}
}
