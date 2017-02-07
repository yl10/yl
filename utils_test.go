package main

import "testing"

func TestGonicCasedName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "UserID",
			args: args{
				name: "UserID",
			},
			want: "user_id",
		},
		{
			name: "UserId",
			args: args{
				name: "UserId",
			},
			want: "user_id",
		},
		{
			name: "UserGUID",
			args: args{
				name: "UserGUID",
			},
			want: "user_guid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GonicCasedName(tt.args.name); got != tt.want {
				t.Errorf("GonicCasedName() = %v, want %v", got, tt.want)
			}
		})
	}
}
