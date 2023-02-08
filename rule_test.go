package gorule

import (
	"errors"
	"reflect"
	"testing"
)

func TestRule_Execute(t *testing.T) {
	type fields struct {
		name      string
		condition string
		action    func(interface{}) (interface{}, error)
	}
	type args struct {
		input interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				name:      "test_rule",
				condition: "1 = 1",
				action: func(i interface{}) (interface{}, error) {
					return nil, errors.New("failed")
				},
			},
			args:    args{input: "input"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Rule{
				name:      tt.fields.name,
				condition: tt.fields.condition,
				action:    tt.fields.action,
			}
			got, err := r.Execute(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Execute() got = %v, want %v", got, tt.want)
			}
		})
	}
}
