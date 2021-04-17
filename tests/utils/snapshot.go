package testutils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"
)

// MustWriteSnapshot updates the snapshot file for a given test t.
func MustWriteSnapshot(t *testing.T, name string, content []byte) {
	Must(t, ioutil.WriteFile(name, content, 0644), "failed to update snapshot", name)
}

// MustReadSnapshot reads the snapshot file for a given test t.
func MustReadSnapshot(t *testing.T, name string) (content []byte) {
	g, err := ioutil.ReadFile(name)
	Must(t, err, "failed reading file", name)
	return g
}

// MustCheckSnapshot compares v with its snapshot file
func MustCheckSnapshot(t *testing.T, name string, v interface{}) {
	var err error
	res, ok := v.([]byte)
	if !ok {
		res, err = json.Marshal(v)
		Must(t, err)
	}
	if *UpdateFlag {
		MustWriteSnapshot(t, name, res)
	}
	g := MustReadSnapshot(t, name)
	if !bytes.Equal(res, g) {
		t.Fatalf("err: %s not equal to %s", res, g)
	}
}
