package utils

import "testing"

func TestGetAdminHost(t *testing.T) {
	tcp := "127.0.0.1:8000"

	admin := GetAdminHost(tcp)

	if admin != "127.0.0.1:8001" {
		t.Errorf("TestGetAdminHost Got %s, expected 127.0.0.1:8001", admin)
	}

}
