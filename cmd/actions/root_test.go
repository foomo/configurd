package actions

import (
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
)

func Test_newLogger(t *testing.T) {
	type args struct {
		verbose bool
	}
	tests := []struct {
		name      string
		args      args
		wantLevel logrus.Level
	}{
		{"standard", args{verbose: false}, logrus.InfoLevel},
		{"verbose", args{verbose: true}, logrus.TraceLevel},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newLogger(tt.args.verbose).Level; !reflect.DeepEqual(got, tt.wantLevel) {
				t.Errorf("newLogger() = %v, want %v", got, tt.wantLevel)
			}
		})
	}
}
