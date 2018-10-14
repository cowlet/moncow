package parser

import (
	"fmt"
	"github.com/cowlet/moncow/ast"
	"github.com/cowlet/moncow/lexer"
	"github.com/cowlet/moncow/token"
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // >, <
	SUM         // +, -
	PRODUCT     // *, /
	PREFIX      // -x, !x
	CALL        // fn(x)
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l            *lexer.Lexer
	currentToken token.Token
	peekToken    token.Token
	errors       []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}
	/* Read two tokens into current and peek */
	p.nextToken()
	p.nextToken()

	/* Set up operator functions */
	p.prefixParseFns = map[token.TokenType]prefixParseFn{
		token.IDENT: p.parseIdentifier,
	}

	return p
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) tokenError(t token.TokenType) {
	msg := fmt.Sprintf("Expected token type %s, got %s instead",
		t, p.currentToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) validateToken(t token.TokenType) (token.Token, bool) {
	tok := p.currentToken
	ok := true
	if p.currentToken.Type != t {
		p.tokenError(t)
		ok = false
	}
	p.nextToken()
	return tok, ok
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	/* Expect LET, IDENT, ASSIGN, <expression>, SEMI */
	let, ok := p.validateToken(token.LET)
	if !ok {
		return nil
	}
	name, ok := p.validateToken(token.IDENT)
	if !ok {
		return nil
	}
	ident := &ast.Identifier{Token: name, Value: name.Literal}
	_, ok = p.validateToken(token.ASSIGN)
	if !ok {
		return nil
	}

	/* TODO: revisit */
	for p.currentToken.Type != token.SEMI {
		p.nextToken()
	}
	//expression := p.parseExpression()
	return &ast.LetStatement{Token: let, Name: ident} // skip Value
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	/* Expect RETURN, <expression>, SEMI */
	ret, ok := p.validateToken(token.RETURN)
	if !ok {
		return nil
	}

	/* TODO: revisit */
	for p.currentToken.Type != token.SEMI {
		p.nextToken()
	}
	//expression := p.parseExpression()
	return &ast.ReturnStatement{Token: ret} // skip Value
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.currentToken}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekToken.Type == token.SEMI {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.currentToken.Type]
	if prefix == nil {
		return nil
	}
	leftExp := prefix()

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
}

func (p *Parser) parseStatement() ast.Statement {
	fmt.Printf("Parsing statement beginning '%s'\n", p.currentToken.Type)
	switch p.currentToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.currentToken.Type != token.EOF {
		statement := p.parseStatement()

		if statement != nil {
			fmt.Printf("Parsed statement %q\n", statement.String())
			program.Statements = append(program.Statements, statement)
		}
		p.nextToken()
	}
	return program
}
