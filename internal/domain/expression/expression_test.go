package expression

import (
	"reflect"
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalculate(t *testing.T) {
	type args struct {
		expression string
		amount     map[string]decimal.Decimal
		fee        map[string]decimal.Decimal
	}
	tests := []struct {
		name    string
		args    args
		want    decimal.Decimal
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Calculate(tt.args.expression, tt.args.amount, tt.args.fee)
			if (err != nil) != tt.wantErr {
				t.Errorf("Calculate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Calculate() = %v, want %v", got, tt.want)
			}
		})
	}
}
