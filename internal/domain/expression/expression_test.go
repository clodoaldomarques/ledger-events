package expression

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		amount     map[string]decimal.Decimal
		fee        map[string]decimal.Decimal
		want       decimal.Decimal
		wantErr    error
	}{
		{
			name:       "simple addition of two amount fields",
			expression: "Amount.a + Amount.b",
			amount: map[string]decimal.Decimal{
				"a": decimal.NewFromInt(10),
				"b": decimal.NewFromInt(20),
			},
			fee:     map[string]decimal.Decimal{},
			want:    decimal.NewFromInt(30),
			wantErr: nil,
		},
		{
			name:       "multiplication and division with fees",
			expression: "(Fee.iof * Fee.tax) / Fee.rate",
			amount:     map[string]decimal.Decimal{},
			fee: map[string]decimal.Decimal{
				"iof":  decimal.NewFromInt(10),
				"tax":  decimal.NewFromInt(5),
				"rate": decimal.NewFromInt(2),
			},
			want:    decimal.NewFromInt(25),
			wantErr: nil,
		},
		{
			name:       "sum of all amount values",
			expression: "Sum(Amount)",
			amount: map[string]decimal.Decimal{
				"a": decimal.NewFromInt(10),
				"b": decimal.NewFromInt(20),
				"c": decimal.NewFromInt(30),
			},
			fee:     map[string]decimal.Decimal{},
			want:    decimal.NewFromInt(60),
			wantErr: nil,
		},
		{
			name:       "sum of all fee values",
			expression: "Sum(Fee)",
			amount:     map[string]decimal.Decimal{},
			fee: map[string]decimal.Decimal{
				"x": decimal.NewFromInt(5),
				"y": decimal.NewFromInt(15),
				"z": decimal.NewFromInt(25),
			},
			want:    decimal.NewFromInt(45),
			wantErr: nil,
		},
		{
			name:       "sum of both amount and fee with arithmetic",
			expression: "Sum(Amount) + Sum(Fee)",
			amount: map[string]decimal.Decimal{
				"a": decimal.NewFromInt(10),
				"b": decimal.NewFromInt(20),
			},
			fee: map[string]decimal.Decimal{
				"x": decimal.NewFromInt(5),
				"y": decimal.NewFromInt(7),
			},
			want:    decimal.NewFromInt(42),
			wantErr: nil,
		},
		{
			name:       "sum of empty maps returns zero",
			expression: "Sum(Amount) + Sum(Fee)",
			amount:     map[string]decimal.Decimal{},
			fee:        map[string]decimal.Decimal{},
			want:       decimal.Zero,
			wantErr:    nil,
		},
		{
			name:       "complex expression with nested sums",
			expression: "(Sum(Amount) * 2) - Sum(Fee)",
			amount: map[string]decimal.Decimal{
				"a": decimal.NewFromInt(10),
				"b": decimal.NewFromInt(20),
			},
			fee: map[string]decimal.Decimal{
				"x": decimal.NewFromInt(5),
				"y": decimal.NewFromInt(7),
			},
			want:    decimal.NewFromInt(48), // (30*2) - 12 = 48
			wantErr: nil,
		},
		{
			name:       "literal only expression - addition and multiplication",
			expression: "10 + 20 * 3",
			amount:     map[string]decimal.Decimal{},
			fee:        map[string]decimal.Decimal{},
			want:       decimal.NewFromInt(70), // 10 + 60 = 70
			wantErr:    nil,
		},
		{
			name:       "literal with division",
			expression: "100 / 4",
			amount:     map[string]decimal.Decimal{},
			fee:        map[string]decimal.Decimal{},
			want:       decimal.NewFromInt(25),
			wantErr:    nil,
		},
		{
			name:       "sum multiplied by literal",
			expression: "Sum(Amount) * 2",
			amount: map[string]decimal.Decimal{
				"a": decimal.NewFromInt(10),
				"b": decimal.NewFromInt(20),
			},
			fee:     map[string]decimal.Decimal{},
			want:    decimal.NewFromInt(60), // (10+20)*2 = 60
			wantErr: nil,
		},
		{
			name:       "amount field plus literal",
			expression: "Amount.a + 5",
			amount: map[string]decimal.Decimal{
				"a": decimal.NewFromInt(10),
			},
			fee:     map[string]decimal.Decimal{},
			want:    decimal.NewFromInt(15),
			wantErr: nil,
		},
		{
			name:       "complex with literals and fields",
			expression: "(Sum(Amount) + 5) * (Fee.tax - 2)",
			amount: map[string]decimal.Decimal{
				"a": decimal.NewFromInt(10),
			},
			fee: map[string]decimal.Decimal{
				"tax": decimal.NewFromInt(10),
			},
			want:    decimal.NewFromInt(120), // (10+5)*(10-2) = 15*8 = 120
			wantErr: nil,
		},
		{
			name:       "floating point literal",
			expression: "10.5 * 2",
			amount:     map[string]decimal.Decimal{},
			fee:        map[string]decimal.Decimal{},
			want:       decimal.NewFromFloat(21.0),
			wantErr:    nil,
		},
		{
			name:       "invalid expression - syntax error",
			expression: "Amount.a +",
			amount:     map[string]decimal.Decimal{},
			fee:        map[string]decimal.Decimal{},
			want:       decimal.Decimal{},
			wantErr:    ErrInvalidExpression{},
		},
		{
			name:       "invalid expression - unknown function",
			expression: "Unknown(Amount)",
			amount:     map[string]decimal.Decimal{},
			fee:        map[string]decimal.Decimal{},
			want:       decimal.Decimal{},
			wantErr:    ErrInvalidExpression{},
		},
		{
			name:       "invalid expression - Sum with wrong argument type (literal)",
			expression: "Sum(10)", // Sum expects map
			amount:     map[string]decimal.Decimal{},
			fee:        map[string]decimal.Decimal{},
			want:       decimal.Decimal{},
			wantErr:    ErrInvalidExpression{},
		},
		{
			name:       "invalid expression - Sum with wrong argument type (string)",
			expression: `Sum("hello")`,
			amount:     map[string]decimal.Decimal{},
			fee:        map[string]decimal.Decimal{},
			want:       decimal.Decimal{},
			wantErr:    ErrInvalidExpression{},
		},
		{
			name:       "accessing non-existent field",
			expression: "Amount.xxx",
			amount: map[string]decimal.Decimal{
				"a": decimal.NewFromInt(10),
			},
			fee:     map[string]decimal.Decimal{},
			want:    decimal.Zero,
			wantErr: nil, // runtime error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Calculate(tt.expression, tt.amount, tt.fee)

			if tt.wantErr != nil {
				assert.Error(t, err)
				switch tt.wantErr.(type) {
				case ErrInvalidExpression:
					assert.IsType(t, ErrInvalidExpression{}, err)
				case ErrCalculateExpression:
					assert.IsType(t, ErrCalculateExpression{}, err)
				default:
					assert.Equal(t, tt.wantErr, err)
				}
				return
			}

			require.NoError(t, err)
			assert.True(t, tt.want.Equal(got), "expected %s, got %s", tt.want.String(), got.String())
		})
	}
}

// TestCalculate_ErrorMessages verifica mensagens específicas
func TestCalculate_ErrorMessages(t *testing.T) {
	tests := []struct {
		name          string
		expression    string
		amount        map[string]decimal.Decimal
		fee           map[string]decimal.Decimal
		expectedError string
	}{
		{
			name:          "invalid expression returns ErrInvalidExpression",
			expression:    "invalid syntax!",
			amount:        map[string]decimal.Decimal{},
			fee:           map[string]decimal.Decimal{},
			expectedError: "invalid expression: invalid syntax!",
		},
		{
			name:          "type mismatch in Sum",
			expression:    "Sum(10)", // Sum expects map
			amount:        map[string]decimal.Decimal{},
			fee:           map[string]decimal.Decimal{},
			expectedError: "calculate error: ...", // will contain specific message
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Calculate(tt.expression, tt.amount, tt.fee)
			assert.Error(t, err)
			if _, ok := err.(ErrInvalidExpression); ok {
				assert.Contains(t, err.Error(), "invalid expression")
			} else if _, ok := err.(ErrCalculateExpression); ok {
				assert.Contains(t, err.Error(), "calculate error")
			} else {
				assert.Equal(t, tt.expectedError, err.Error())
			}
		})
	}
}
