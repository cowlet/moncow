package parser

import (
	"fmt"
	"github.com/cowlet/moncow/ast"
	"github.com/cowlet/moncow/lexer"
	"github.com/cowlet/moncow/token"
	"testing"
)

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func initParser(t *testing.T, input string, numStatements int) *ast.Program {
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != numStatements {
		t.Fatalf("program.Statements doesn't contain %d (got %d):\n%s", numStatements, len(program.Statements), program.String())
	}

	return program
}

func TestLetStatements(t *testing.T) {
	input := `
let x = 5;
let y = 10.3;
let moo = 12345;
`
	program := initParser(t, input, 3)

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"moo"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral() not 'let'. Got '%q'", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("statement not *ast.LetStatement. Got %T", s)
		return false
	}

	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. Got '%s'", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("letStmt.Name.TokenLiteral() not '%s'. Got '%s'", name, letStmt.Name.TokenLiteral())
		return false
	}

	return true
}

func TestReturnStatements(t *testing.T) {
	input := `
return 5;
return 10.5;
return 12345;
`
	program := initParser(t, input, 3)

	for _, stmt := range program.Statements {
		retStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("statement not *ast.ReturnStatement. Got %T", retStmt)
			continue
		}
		if stmt.TokenLiteral() != "return" {
			t.Errorf("stmt.TokenLiteral not 'return', got %q", stmt.TokenLiteral())
		}
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "moo;"
	program := initParser(t, input, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement not ast.ExpressionStatement. Got %T", program.Statements[0])
	}

	exp, ok := stmt.Expression.(ast.Expression)
	if !testIdentifier(t, exp, "moo") {
		return
	}
}

func TestBooleanExpression(t *testing.T) {
	input := `
true;
false;
`
	program := initParser(t, input, 2) /* 2 statements */

	tests := []struct {
		expected bool
	}{
		{true},
		{false},
	}

	for i, tt := range tests {
		stmt, ok := program.Statements[i].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("statement not *ast.ExpressionStatement. Got %T", stmt)
			continue
		}
		if !testBoolean(t, stmt.Expression, tt.expected) {
			return
		}
	}
}

func TestNumericLiteralExpressions(t *testing.T) {
	input := `
5;
10.5;
`
	program := initParser(t, input, 2)

	/* Integer */
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement not ast.ExpressionStatement. Got %T", program.Statements[0])
	}
	if !testIntegerLiteral(t, stmt.Expression, 5) {
		return
	}

	/* Float */
	stmt, ok = program.Statements[1].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement not ast.ExpressionStatement. Got %T", program.Statements[0])
	}
	if !testFloatLiteral(t, stmt.Expression, 10.5) {
		return
	}
}

/* Helper functions for expression parsing */
func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. Got %T", il)
		return false
	}

	if integ.Value != value {
		t.Errorf("integ.Value not %d. Got %d", value, integ.Value)
		return false
	}
	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d. Got %s", value, integ.TokenLiteral())
		return false
	}
	return true
}

func testFloatLiteral(t *testing.T, fl ast.Expression, value float64) bool {
	flt, ok := fl.(*ast.FloatLiteral)
	if !ok {
		t.Errorf("fl not *ast.FloatLiteral. Got %T", fl)
		return false
	}

	if flt.Value != value {
		t.Errorf("flt.Value not %f. Got %f", value, flt.Value)
		return false
	}
	if flt.TokenLiteral() != fmt.Sprintf("%.1f", value) {
		t.Errorf("flt.TokenLiteral not %f. Got %s", value, flt.TokenLiteral())
		return false
	}
	return true
}

func testIdentifier(t *testing.T, id ast.Expression, value string) bool {
	ident, ok := id.(*ast.Identifier)
	if !ok {
		t.Errorf("id not *ast.Identifier. Got %T", id)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value not %q. Got %q", value, ident.Value)
		return false
	}
	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not %s. Got %s", value, ident.TokenLiteral())
		return false
	}
	return true
}

func testBoolean(t *testing.T, bexp ast.Expression, value bool) bool {
	b, ok := bexp.(*ast.Boolean)
	if !ok {
		t.Errorf("bexp not *ast.Boolean. Got %T", b)
		return false
	}
	if b.TokenLiteral() != fmt.Sprintf("%v", value) {
		t.Errorf("b.TokenLiteral not '%v', got %q", value, b.TokenLiteral())
		return false
	}
	if b.Value != value {
		t.Errorf("b.Value not equal to %v, got %v", value, b.Value)
		return false
	}
	return true
}

func testLiteralExpression(
	t *testing.T, exp ast.Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case float64:
		return testFloatLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBoolean(t, exp, v)
	}
	t.Errorf("type %T not handled", exp)
	return false
}

func testInfixExpression(
	t *testing.T, exp ast.Expression, left interface{},
	operator string, right interface{}) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp not *ast.InfixExpression. Got %T", exp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}
	if opExp.Operator != operator {
		t.Errorf("exp.Operator not '%q'. Got '%q'", operator, opExp.Operator)
		return false
	}
	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}
	return true
}

/* Testing of prefix/infix expressions */
func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"-5.0;", "-", 5.0},
		{"!15.1;", "!", 15.1},
		{"-moo;", "-", "moo"},
	}

	for _, tt := range prefixTests {
		program := initParser(t, tt.input, 1) // 1 statement

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("statement not ast.ExpressionStatement. Got %T", program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("exp not *ast.PrefixExpression. Got %T", stmt.Expression)
		}
		if exp.Operator != tt.operator {
			t.Errorf("exp.Operator not %q. Got %v", tt.operator, exp.Operator)
		}
		if !testLiteralExpression(t, exp.Right, tt.value) {
			return
		}
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"5.8 + 1.2;", 5.8, "+", 1.2},
		{"5 + 5.4;", 5, "+", 5.4},
		{"moo + hoof;", "moo", "+", "hoof"},
		{"true != false;", true, "!=", false},
	}

	for _, tt := range infixTests {
		program := initParser(t, tt.input, 1) // 1 statement

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("statement not ast.ExpressionStatement. Got %T", program.Statements[0])
		}

		if !testInfixExpression(t, stmt.Expression,
			tt.leftValue, tt.operator, tt.rightValue) {
			return
		}
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a)*b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a+b)+c)",
		},
		{
			"a + b - c",
			"((a+b)-c)",
		},
		{
			"a * b * c",
			"((a*b)*c)",
		},
		{
			"a * b / c",
			"((a*b)/c)",
		},
		{
			"a + b / c",
			"(a+(b/c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a+(b*c))+(d/e))-f)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5>4)==(3<4))",
		},
		{
			"5 > 4 != 3 < 4",
			"((5>4)!=(3<4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3+(4*5))==((3*1)+(4*5)))",
		},
		{
			"true == false",
			"(true==false)",
		},
		{
			"a > 5 == false",
			"((a>5)==false)",
		},
		{
			"1 + (2 + 3) + 4",
			"((1+(2+3))+4)",
		},
		{
			"(5 + 5) * 2",
			"((5+5)*2)",
		},
		{
			"-(5 + 5)",
			"(-(5+5))",
		},
		{
			"!(true == true)",
			"(!(true==true))",
		},
	}

	for _, tt := range tests {
		program := initParser(t, tt.input, 1) // 1 statement

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, actual)
		}
	}

	two := "3 + 4; -5 * 5"
	expected := "(3+4)((-5)*5)"
	program := initParser(t, two, 2) // 2 statements
	actual := program.String()
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func testBlock(t *testing.T, blk *ast.BlockStatement, stmts []ast.Statement) bool {
	if blk == nil && stmts == nil {
		return true
	}
	if len(blk.Statements) != len(stmts) {
		t.Errorf("Expected %d statements in block, got %d",
			len(stmts), len(blk.Statements))
		return false
	}
	for i, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.LetStatement:
			actual, ok := blk.Statements[i].(*ast.LetStatement)
			if !ok {
				t.Errorf("Statement %d expected to be of type %T, got %T",
					i, s, blk.Statements[i])
				return false
			}
			if actual.Token.Literal != s.Token.Literal {
				t.Errorf("Expected LetStatement Token to be %s, got %s",
					s.Token.Literal, actual.Token.Literal)
				return false
			}
			if actual.Name.Value != s.Name.Value {
				t.Errorf("Expected LetStatement Name to be %s, got %s",
					s.Name.Value, actual.Name.Value)
				return false
			}
			if actual.Value.String() != s.Value.String() {
				t.Errorf("Expected LetStatement Value to be %s, got %s",
					s.Value.String(), actual.Value.String())
				return false
			}
		case *ast.ReturnStatement:
			actual, ok := blk.Statements[i].(*ast.ReturnStatement)
			if !ok {
				t.Errorf("Statement %d expected to be of type %T, got %T",
					i, s, blk.Statements[i])
				return false
			}
			if actual.Token.Literal != s.Token.Literal {
				t.Errorf("Expected ReturnStatement Token to be %s, got %s",
					s.Token.Literal, actual.Token.Literal)
				return false
			}
			if actual.Value.String() != s.Value.String() {
				t.Errorf("Expected ReturnStatement Value to be %s, got %s",
					s.Value.String(), actual.Value.String())
				return false
			}
		case *ast.ExpressionStatement:
			actual, ok := blk.Statements[i].(*ast.ExpressionStatement)
			if !ok {
				t.Errorf("Statement %d expected to be of type %T, got %T",
					i, s, blk.Statements[i])
				return false
			}
			if actual.String() != s.String() {
				t.Errorf("Expected ExpressionStatement to be %s, got %s",
					s.String(), actual.String())
				return false
			}
		default:
			t.Errorf("Unknown type of Statement %T", s)
			return false
		}
	}
	return true
}

func TestIfExpression(t *testing.T) {
	tokx := token.Token{token.IDENT, "x"}
	toky := token.Token{token.IDENT, "y"}
	expx := &ast.ExpressionStatement{tokx, &ast.Identifier{tokx, "x"}}
	expy := &ast.ExpressionStatement{toky, &ast.Identifier{toky, "y"}}

	tests := []struct {
		input     string
		condition string
		ifblk     []ast.Statement
		elseblk   []ast.Statement
	}{
		{
			"if (x < y) { x }",
			"(x<y)",
			[]ast.Statement{expx},
			nil,
		},
		{
			"if (x < y) { x; y; }",
			"(x<y)",
			[]ast.Statement{expx, expy},
			nil,
		},
		{
			"if (x > y) { x } else { y }",
			"(x>y)",
			[]ast.Statement{expx},
			[]ast.Statement{expy},
		},
	}

	for _, tt := range tests {
		program := initParser(t, tt.input, 1) // 1 statement
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("statement not ast.ExpressionStatement. Got %T",
				program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.IfExpression)
		if !ok {
			t.Fatalf("stmt.Expression is not ast.IfExpression. Got %T",
				stmt.Expression)
		}

		if exp.Condition.String() != tt.condition {
			t.Errorf("exp.Condition not '%q'. Got '%q'",
				exp.Condition.String(), tt.condition)
			return
		}

		if !testBlock(t, exp.IfBlock, tt.ifblk) {
			return
		}

		if !testBlock(t, exp.ElseBlock, tt.elseblk) {
			return
		}
	}

}
