package parser

import (
	"fmt"
	"github.com/cowlet/moncow/ast"
	"github.com/cowlet/moncow/lexer"
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

	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. Got %T", stmt.Expression)
	}
	if ident.Value != "moo" {
		t.Errorf("ident.Value not %s. Got %s", "moo", ident.Value)
	}
	if ident.TokenLiteral() != "moo" {
		t.Errorf("ident.TokenLiteral not %s. Got %s", "moo", ident.TokenLiteral())
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

	exp, ok := stmt.Expression.(ast.Expression)
	if !testIntegerLiteral(t, exp, 5) {
		return
	}

	/* Float */
	stmt, ok = program.Statements[1].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement not ast.ExpressionStatement. Got %T", program.Statements[0])
	}

	lit, ok := stmt.Expression.(*ast.FloatLiteral)
	if !ok {
		t.Fatalf("exp not *ast.FloatLiteral. Got %T", stmt.Expression)
	}
	if lit.Value != 10.5 {
		t.Errorf("lit.Value not 10.5. Got %v", lit.Value)
	}
	if lit.TokenLiteral() != "10.5" {
		t.Errorf("lit.TokenLiteral not '10.5'. Got %s", lit.TokenLiteral())
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input        string
		operator     string
		integerValue int64
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
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
		if !testIntegerLiteral(t, exp.Right, tt.integerValue) {
			return
		}
	}
}

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

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  int64
		operator   string
		rightValue int64
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
	}

	for _, tt := range infixTests {
		program := initParser(t, tt.input, 1) // 1 statement

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("statement not ast.ExpressionStatement. Got %T", program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("exp not *ast.InfixExpression. Got %T", stmt.Expression)
		}
		if !testIntegerLiteral(t, exp.Left, tt.leftValue) {
			return
		}
		if exp.Operator != tt.operator {
			t.Errorf("exp.Operator not %q. Got %v", tt.operator, exp.Operator)
		}
		if !testIntegerLiteral(t, exp.Right, tt.rightValue) {
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
