package service

import "testing"

func Test_isOrderNumberValid(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "succes",
			number: "5062821234567892",
			want:   true,
		},
		{
			name:   "negative_invalid_format",
			number: "50628212 34567892",
			want:   false,
		},
		{
			name:   "negative_Luhn_algorithm",
			number: "5062821234567899",
			want:   false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := isOrderNumberValid(test.number)
			if test.want != got {
				t.Errorf("isOrderNumberValid() = %v, want %v", got, test.want)
			}
		})
	}
}
