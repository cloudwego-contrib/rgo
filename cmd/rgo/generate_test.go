package main

import "testing"

func TestGenerate(t *testing.T) {
	err := GenerateRGOCode()
	if err != nil {
		return
	}
}
