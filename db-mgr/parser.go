package main

import (
	"fmt"
	"strings"
)

type AST struct {
	name   string
	tokens []string
}

var ASTs = []AST{
	{name: "SELECT", tokens: []string{"SELECT", "[fields]", "FROM", "[tablename]", "WHERE", "[conditions]"}},
	{name: "SELECT", tokens: []string{"SELECT", "[fields]", "FROM", "[tablename]"}},
	{name: "INSERT", tokens: []string{"INSERT", "INTO", "[tablename]", "(", "[fields]", ")", "VALUES", "(", "[values]", ")"}},
	{name: "UPDATE", tokens: []string{"UPDATE", "[tablename]", "SET", "[conditions]", "WHERE", "[conditions]"}},
	{name: "fields", tokens: []string{"name", ",", "[fields]"}},
	{name: "fields", tokens: []string{"name"}},
	{name: "function", tokens: []string{"NOW"}},
	{name: "function", tokens: []string{"DATETIME"}},
	{name: "value", tokens: []string{"name"}},
	{name: "value", tokens: []string{"string"}},
	{name: "value", tokens: []string{"[function]", "(", ")"}},
	{name: "values", tokens: []string{"[value]", ",", "[values]"}},
	{name: "values", tokens: []string{"[value]"}},
	{name: "tablename", tokens: []string{"string"}},
	{name: "tablename", tokens: []string{"name"}},
	{name: "condition", tokens: []string{"name", "op", "[value]"}},
	{name: "conditions", tokens: []string{"[condition]", "AND", "[conditions]"}},
	{name: "conditions", tokens: []string{"[condition]"}},
}

type Token struct {
	code     string
	value    string
	children []Token
}

func (token Token) generateSql(database string) string {
	if token.code == "[tablename]" {
		return database + "." + token.children[0].value
	}

	if token.children != nil && len(token.children) > 0 {
		return strings.Join(Map(token.children, func(token Token) string { return token.generateSql(database) }), " ")
	}

	if token.code == "string" {
		return "'" + token.value + "'"
	}

	if token.value != "" {
		return token.value
	}

	return token.code
}

func tokensMatchAst(astTokens []string, tokens []Token, trail []Token) (bool, []Token, []Token) {
	if len(astTokens) == 0 {
		return true, tokens, trail
	}

	if len(tokens) == 0 {
		return false, tokens, trail
	}

	//	fmt.Printf("%s = %s ?\n", astTokens[0], tokens[0].code)

	var remainingTokens []Token
	var groupToken Token
	var isMatch bool

	if strings.HasPrefix(astTokens[0], "[") && strings.HasSuffix(astTokens[0], "]") {
		recurstAstName := astTokens[0][1 : len(astTokens[0])-1]
		isMatch, remainingTokens, groupToken = parseAst(recurstAstName, tokens)
		if !isMatch {
			return false, remainingTokens, trail
		}
		trail = append(trail, groupToken)

	} else if astTokens[0] != tokens[0].code {
		return false, tokens, trail
	} else {
		remainingTokens = tokens[1:]
		trail = append(trail, tokens[0])
	}

	return tokensMatchAst(astTokens[1:], remainingTokens, trail)
}

func parseAst(astName string, tokens []Token) (bool, []Token, Token) {
	for _, ast := range ASTs {
		if ast.name != astName {
			continue
		}
		/*
			if id == recursAst {
				continue
			}
		*/
		isMatch, tokensLeft, trail := tokensMatchAst(ast.tokens, tokens, make([]Token, 0))
		if isMatch {
			return true, tokensLeft, Token{code: "[" + astName + "]", children: trail}
		}
	}

	return false, tokens, Token{}
}

func flatten(groupCode string, code string, separatorCode string, tokens []Token) []Token {
	if len(tokens) == 1 && tokens[0].code == code {
		return tokens
	}

	if len(tokens) != 3 || tokens[0].code != code || tokens[1].code != separatorCode || tokens[2].code != groupCode {
		return make([]Token, 0)
	}
	return append([]Token{tokens[0]}, flatten(groupCode, code, separatorCode, tokens[2].children)...)
}

func (parentToken Token) reduce(groupCode string, code string, separatorCode string) {
	for i, token := range parentToken.children {
		if token.code == groupCode {
			token.children = flatten(groupCode, code, separatorCode, token.children)
			parentToken.children[i] = token
		}
	}
}

func ParseSql(sql string, database string) (string, error) {
	tokens := []Token{}
	var token Token
	var err error
	var res bool

	chars := []rune(sql)

	pos := 0
	for pos >= 0 {
		token, pos, err = nextToken(chars, pos)
		if err != nil {
			return "", err
		}
		if pos >= 0 {
			tokens = append(tokens, token)
		}
	}

	for _, token := range tokens {
		if token.value == "" {
			fmt.Printf("[%s] ", token.code)
		} else {
			fmt.Printf("[%s|%s] ", token.code, token.value)
		}
	}

	switch tokens[0].code {
	case "SELECT", "INSERT", "UPDATE", "DELETE":
		res, _, token = parseAst(tokens[0].code, tokens)
	default:
		res = false
	}

	if res {
		fmt.Println("Match!")
		return token.generateSql(database), nil
		//		token.reduce("[conditions]", "[condition]", "AND")
		//		token.reduce("[fields]", "name", ",")
		//		token.reduce("[values]", "[value]", ",")
	} else {
		fmt.Println("No match")
		return "", fmt.Errorf("Cannot parse SQL query")
	}
}

func nextToken(chars []rune, pos int) (Token, int, error) {
	end := len(chars)

	c := ' '
	for pos < end {
		c = chars[pos]
		if c == ' ' || c == '\n' || c == '\r' || c == '\t' {
			pos += 1
		} else {
			break
		}
	}

	if pos == end {
		return Token{}, -1, nil
	}

	startPos := pos

	var quote rune

	c = chars[pos]

	switch c {
	case '(', ')', ',', '.':
		return Token{code: string(c)}, pos + 1, nil
	case '\'', '"':
		quote = c
		pos += 1
		for pos < end {
			c = chars[pos]
			if c == quote {
				return Token{code: "string", value: string(chars[startPos+1 : pos])}, pos + 1, nil
			}
			if c == '\\' {
				pos += 1
			}
			pos += 1
		}

		return Token{}, -1, fmt.Errorf("Unclosed quote " + string(quote))
	case '=', '!', '<', '>':
		pos += 1
		for pos < end {
			c = chars[pos]
			if c == '=' || c == '!' || c == '<' || c == '>' {
				pos += 1
			} else {
				break
			}
		}

		value := string(chars[startPos:pos])
		switch value {
		case "=", "<", ">", "<=", ">=", "!=", "<>":
			return Token{code: "op", value: value}, pos + 1, nil
		default:
			return Token{}, -1, fmt.Errorf("Unknown operator: [" + value + "]")
		}
	}

	pos += 1
	for pos < end {
		c = chars[pos]

		if (c >= 'a' && c <= 'z') || (c == '_') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			pos += 1
		} else {
			break
		}
	}

	value := string(chars[startPos:pos])

	switch strings.ToUpper(value) {
	case "SELECT", "FROM", "WHERE", "AND", "INSERT", "INTO", "VALUES", "NOW", "UPDATE", "SET":
		return Token{code: value}, pos, nil
	default:
		return Token{code: "name", value: value}, pos, nil
	}
}
