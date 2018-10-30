package parser

import (
	"fmt"
	"github.com/cowlet/moncow/ast"
	"github.com/cowlet/moncow/lexer"
	"github.com/cowlet/moncow/token"
	"strconv"
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

var precedences = map[token.TokenType]int{
	token.EQ:    EQUALS,
	token.NEQ:   EQUALS,
	token.LT:    LESSGREATER,
	token.GT:    LESSGREATER,
	token.PLUS:  SUM,
	token.MINUS: SUM,
	token.MULT:  PRODUCT,
	token.DIV:   PRODUCT,
}

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
		token.INT:   p.parseIntegerLiteral,
		token.FLOAT: p.parseFloatLiteral,
		token.BANG:  p.parsePrefixExpression,
		token.MINUS: p.parsePrefixExpression,
	}

	p.infixParseFns = map[token.TokenType]infixParseFn{
		token.PLUS:  p.parseInfixExpression,
		token.MINUS: p.parseInfixExpression,
		token.MULT:  p.parseInfixExpression,
		token.DIV:   p.parseInfixExpression,
		token.EQ:    p.parseInfixExpression,
		token.NEQ:   p.parseInfixExpression,
		token.GT:    p.parseInfixExpression,
		token.LT:    p.parseInfixExpression,
	}

	return p
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) precedence(tt token.Token) int {
	if p, ok := precedences[tt.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) tokenError(t token.TokenType) {
	msg := fmt.Sprintf("Expected token type %s, got %s instead",
		t, p.currentToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("No prefix parse function for %s found", t)
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
		p.noPrefixParseFnError(p.currentToken.Type)
		return nil
	}
	leftExp := prefix()

	for p.peekToken.Type != token.SEMI && precedence < p.precedence(p.peekToken) {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.currentToken}

	value, err := strconv.ParseInt(p.currentToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("Could not parse %q as integer", p.currentToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.currentToken}

	value, err := strconv.ParseFloat(p.currentToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("Could not parse %q as float", p.currentToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	pe := &ast.PrefixExpression{
		Token:    p.currentToken,
		Operator: p.currentToken.Literal,
	}
	p.nextToken()
	pe.Right = p.parseExpression(PREFIX)

	return pe
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	ie := &ast.InfixExpression{
		Token:    p.currentToken,
		Operator: p.currentToken.Literal,
		Left:     left,
	}
	precedence := p.precedence(p.currentToken)
	p.nextToken()
	ie.Right = p.parseExpression(precedence)

	return ie
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
