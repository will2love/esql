package esql

import (
	"fmt"

	"github.com/xwb1989/sqlparser"
)

func (e *ESql) convertToScript(expr sqlparser.Expr) (script string, aggFuncSlice []*sqlparser.FuncExpr, aggFuncNameSlice []string, err error) {
	switch expr.(type) {
	case *sqlparser.ColName:
		exprColName := expr.(*sqlparser.ColName)
		script, err = e.convertColName(exprColName)
		script = fmt.Sprintf(`doc['%v'].value`, script)
	case *sqlparser.SQLVal:
		script, err = e.convertValExpr(expr, true)
	case *sqlparser.BinaryExpr:
		script, aggFuncSlice, aggFuncNameSlice, err = e.convertBinaryExpr(expr)
	case *sqlparser.ParenExpr:
		parenExpr := expr.(*sqlparser.ParenExpr)
		script, aggFuncSlice, aggFuncNameSlice, err = e.convertToScript(parenExpr.Expr)
		script = fmt.Sprintf(`(%v)`, script)
	case *sqlparser.UnaryExpr:
		script, aggFuncSlice, aggFuncNameSlice, err = e.convertUnaryExpr(expr)
	case *sqlparser.FuncExpr:
		aggFuncSlice = append(aggFuncSlice, expr.(*sqlparser.FuncExpr))
		aggFuncNameSlice = append(aggFuncNameSlice, sqlparser.String(expr.(*sqlparser.FuncExpr).Exprs))
		script, err = e.convertFuncExpr(expr)
	default:
		err = fmt.Errorf("esql: invalid expression type for scripting")
	}
	if err != nil {
		return "", nil, nil, err
	}
	return script, aggFuncSlice, aggFuncNameSlice, nil
}

func (e *ESql) convertFuncExpr(expr sqlparser.Expr) (string, error) {

	return "", nil
}

func (e *ESql) convertUnaryExpr(expr sqlparser.Expr) (script string, aggFuncSlice []*sqlparser.FuncExpr, aggFuncNameSlice []string, err error) {
	var expStr string
	unaryExpr, ok := expr.(*sqlparser.UnaryExpr)
	if !ok {
		err = fmt.Errorf("esql: invalid unary expression")
		return "", nil, nil, err
	}
	op, ok := opUnaryExpr[unaryExpr.Operator]
	if !ok {
		err = fmt.Errorf("esql: not supported binary expression operator")
		return "", nil, nil, err
	}
	expStr, aggFuncSlice, aggFuncNameSlice, err = e.convertToScript(unaryExpr.Expr)
	if err != nil {
		return "", nil, nil, err
	}

	script = fmt.Sprintf(`%v%v`, op, expStr)
	return script, aggFuncSlice, aggFuncNameSlice, nil
}

func (e *ESql) convertBinaryExpr(expr sqlparser.Expr) (script string, aggFuncSlice []*sqlparser.FuncExpr, aggFuncNameSlice []string, err error) {
	var lhsStr, rhsStr string
	binExpr, ok := expr.(*sqlparser.BinaryExpr)
	if !ok {
		err = fmt.Errorf("esql: invalid binary expression")
		return "", nil, nil, err
	}
	lhsExpr, rhsExpr := binExpr.Left, binExpr.Right
	op, ok := opBinaryExpr[binExpr.Operator]
	if !ok {
		err = fmt.Errorf("esql: not supported binary expression operator")
		return "", nil, nil, err
	}

	var aggFuncs []*sqlparser.FuncExpr
	var aggNames []string
	lhsStr, aggFuncSlice, aggFuncNameSlice, err = e.convertToScript(lhsExpr)
	if err != nil {
		return "", nil, nil, err
	}
	rhsStr, aggFuncs, aggNames, err = e.convertToScript(rhsExpr)
	if err != nil {
		return "", nil, nil, err
	}
	aggFuncNameSlice = append(aggFuncNameSlice, aggNames...)
	aggFuncSlice = append(aggFuncSlice, aggFuncs...)

	script = fmt.Sprintf(`%v %v %v`, lhsStr, op, rhsStr)
	return script, aggFuncSlice, aggFuncNameSlice, nil
}
