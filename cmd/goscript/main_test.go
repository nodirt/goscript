package main

import (
	"testing"
)

func TestIgnoredError(t *testing.T) {
	err := Run([]string{"testdata/ignoredError.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
}
