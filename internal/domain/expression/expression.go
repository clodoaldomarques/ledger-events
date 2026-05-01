package expression

import (
	"github.com/expr-lang/expr"
	"github.com/shopspring/decimal"
)

type Env struct {
	Amount map[string]decimal.Decimal
	Fee    map[string]decimal.Decimal
	Sub    func(a, b decimal.Decimal) decimal.Decimal
	Add    func(a, b decimal.Decimal) decimal.Decimal
	Mul    func(a, b decimal.Decimal) decimal.Decimal
	Div    func(a, b decimal.Decimal) decimal.Decimal
}

func Calculate(expression string, amount, fee map[string]decimal.Decimal) (decimal.Decimal, error) {
	options := []expr.Option{
		expr.Env(Env{}),
		expr.Operator("+", "Add"),
		expr.Operator("-", "Sub"),
		expr.Operator("*", "Mul"),
		expr.Operator("/", "Div"),
	}

	program, err := expr.Compile(expression, options...)
	if err != nil {
		return decimal.Decimal{}, ErrInvalidExpression{expression}
	}

	env := Env{
		Amount: amount,
		Fee:    fee,
		Sub:    func(a, b decimal.Decimal) decimal.Decimal { return a.Sub(b) },
		Add:    func(a, b decimal.Decimal) decimal.Decimal { return a.Add(b) },
		Mul:    func(a, b decimal.Decimal) decimal.Decimal { return a.Mul(b) },
		Div:    func(a, b decimal.Decimal) decimal.Decimal { return a.Div(b) },
	}

	output, err := expr.Run(program, env)
	if err != nil {
		return decimal.Decimal{}, ErrCalculateExpression{err.Error()}
	}

	return output.(decimal.Decimal), nil
}
