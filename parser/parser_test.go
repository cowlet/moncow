package parser

import (
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
		t.Fatalf("program.Statements doesn't contain %d (got %d)", numStatements, len(program.Statements))
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

	intLit, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. Got %T", stmt.Expression)
	}
	if intLit.Value != 5 {
		t.Errorf("intLit.Value not 5. Got %v", intLit.Value)
	}
	if intLit.TokenLiteral() != "5" {
		t.Errorf("intLit.TokenLiteral not '5'. Got %s", intLit.TokenLiteral())
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
