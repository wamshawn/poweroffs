package service_test

import (
	"testing"

	"github.com/wamshawn/poweroffs/service"
)

func TestGen(t *testing.T) {
	err := service.Gen("/home/radxa/temp/poweroffs")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("success")
}
