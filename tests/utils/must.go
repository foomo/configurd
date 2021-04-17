package testutils

import (
	"fmt"
	"testing"
)

func Must(t *testing.T, err error, args ...interface{}) {
	if err != nil {
		args = append([]interface{}{fmt.Sprintf("err: %s", err)}, args...)
		t.Fatal(args...)
	}
}
