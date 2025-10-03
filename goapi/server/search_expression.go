package server

import (
	"fmt"
	"strings"
	"unicode"
)

type SearchExpression struct {
	root searchNode
}

type searchNode interface {
	eval(idx *searchDocumentIndex) bool
}

type searchTokenKind int

const (
	tokenTerm searchTokenKind = iota
	tokenPhrase
	tokenAnd
	tokenOr
	tokenNot
	tokenLParen
	tokenRParen
)

type searchToken struct {
	kind  searchTokenKind
	value string
}

type logicalOperator int

const (
	logicalAnd logicalOperator = iota
	logicalOr
)

type termNode struct {
	term     string
	isPhrase bool
}

type notNode struct {
	child searchNode
}

type binaryNode struct {
	op    logicalOperator
	left  searchNode
	right searchNode
}

type searchDocumentIndex struct {
	text   string
	tokens map[string]struct{}
}

func (expr *SearchExpression) Evaluate(idx *searchDocumentIndex) bool {
	if expr == nil || expr.root == nil {
		return true
	}
	return expr.root.eval(idx)
}

func ParseSearchQuery(raw string) (*SearchExpression, error) {
	tokens, err := tokenizeSearchQuery(raw)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return &SearchExpression{}, nil
	}

	tokens = insertImplicitAnd(tokens)

	rpn, err := shuntingYard(tokens)
	if err != nil {
		return nil, err
	}

	root, err := buildSearchAST(rpn)
	if err != nil {
		return nil, err
	}

	return &SearchExpression{root: root}, nil
}

func tokenizeSearchQuery(raw string) ([]searchToken, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	runes := []rune(raw)
	tokens := make([]searchToken, 0, len(runes))

	for i := 0; i < len(runes); {
		ch := runes[i]

		switch {
		case unicode.IsSpace(ch):
			i++
		case ch == '(':
			tokens = append(tokens, searchToken{kind: tokenLParen})
			i++
		case ch == ')':
			tokens = append(tokens, searchToken{kind: tokenRParen})
			i++
		case ch == '"':
			i++
			var builder strings.Builder
			for i < len(runes) {
				c := runes[i]
				if c == '"' {
					break
				}
				if c == '\\' && i+1 < len(runes) {
					i++
					c = runes[i]
				}
				builder.WriteRune(c)
				i++
			}
			if i >= len(runes) || runes[i] != '"' {
				return nil, fmt.Errorf("unterminated phrase in search query")
			}
			i++
			phrase := strings.ToLower(strings.TrimSpace(builder.String()))
			if phrase != "" {
				tokens = append(tokens, searchToken{kind: tokenPhrase, value: phrase})
			}
		case ch == '&' && i+1 < len(runes) && runes[i+1] == '&':
			tokens = append(tokens, searchToken{kind: tokenAnd})
			i += 2
		case ch == '|' && i+1 < len(runes) && runes[i+1] == '|':
			tokens = append(tokens, searchToken{kind: tokenOr})
			i += 2
		case ch == '!':
			tokens = append(tokens, searchToken{kind: tokenNot})
			i++
		default:
			start := i
			for i < len(runes) {
				c := runes[i]
				if unicode.IsSpace(c) || c == '(' || c == ')' {
					break
				}
				if c == '"' {
					break
				}
				i++
			}
			if start == i {
				i++
				continue
			}
			word := string(runes[start:i])
			upper := strings.ToUpper(word)
			switch upper {
			case "AND":
				tokens = append(tokens, searchToken{kind: tokenAnd})
			case "OR":
				tokens = append(tokens, searchToken{kind: tokenOr})
			case "NOT":
				tokens = append(tokens, searchToken{kind: tokenNot})
			default:
				term := strings.ToLower(strings.TrimSpace(word))
				if term != "" {
					tokens = append(tokens, searchToken{kind: tokenTerm, value: term})
				}
			}
		}
	}

	return tokens, nil
}

func insertImplicitAnd(tokens []searchToken) []searchToken {
	if len(tokens) < 2 {
		return tokens
	}

	result := make([]searchToken, 0, len(tokens)*2)
	prev := tokens[0]
	result = append(result, prev)

	for i := 1; i < len(tokens); i++ {
		current := tokens[i]
		if needsImplicitAnd(prev, current) {
			result = append(result, searchToken{kind: tokenAnd})
		}
		result = append(result, current)
		prev = current
	}

	return result
}

func needsImplicitAnd(prev, next searchToken) bool {
	isTermLike := func(tok searchToken) bool {
		return tok.kind == tokenTerm || tok.kind == tokenPhrase || tok.kind == tokenRParen
	}

	beginsTerm := func(tok searchToken) bool {
		return tok.kind == tokenTerm || tok.kind == tokenPhrase || tok.kind == tokenLParen || tok.kind == tokenNot
	}

	return isTermLike(prev) && beginsTerm(next)
}

var precedence = map[searchTokenKind]int{
	tokenOr:  1,
	tokenAnd: 2,
	tokenNot: 3,
}

func shuntingYard(tokens []searchToken) ([]searchToken, error) {
	output := make([]searchToken, 0, len(tokens))
	stack := make([]searchToken, 0, len(tokens))

	for idx, tok := range tokens {
		switch tok.kind {
		case tokenTerm, tokenPhrase:
			output = append(output, tok)
		case tokenNot, tokenAnd, tokenOr:
			currPrec := precedence[tok.kind]
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top.kind == tokenLParen {
					break
				}
				topPrec := precedence[top.kind]
				if tok.kind == tokenNot {
					if topPrec > currPrec {
						output = append(output, top)
						stack = stack[:len(stack)-1]
						continue
					}
				} else {
					if topPrec >= currPrec {
						output = append(output, top)
						stack = stack[:len(stack)-1]
						continue
					}
				}
				break
			}
			stack = append(stack, tok)
		case tokenLParen:
			stack = append(stack, tok)
		case tokenRParen:
			matched := false
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if top.kind == tokenLParen {
					matched = true
					break
				}
				output = append(output, top)
			}
			if !matched {
				return nil, fmt.Errorf("unbalanced parentheses near token %d", idx)
			}
		default:
			return nil, fmt.Errorf("unsupported token in search query")
		}
	}

	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if top.kind == tokenLParen || top.kind == tokenRParen {
			return nil, fmt.Errorf("unbalanced parentheses in search query")
		}
		output = append(output, top)
	}

	return output, nil
}

func buildSearchAST(tokens []searchToken) (searchNode, error) {
	stack := make([]searchNode, 0, len(tokens))

	for _, tok := range tokens {
		switch tok.kind {
		case tokenTerm:
			stack = append(stack, &termNode{term: tok.value})
		case tokenPhrase:
			stack = append(stack, &termNode{term: tok.value, isPhrase: true})
		case tokenNot:
			if len(stack) < 1 {
				return nil, fmt.Errorf("NOT operator missing operand")
			}
			operand := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			stack = append(stack, &notNode{child: operand})
		case tokenAnd:
			if len(stack) < 2 {
				return nil, fmt.Errorf("AND operator missing operands")
			}
			right := stack[len(stack)-1]
			left := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, &binaryNode{op: logicalAnd, left: left, right: right})
		case tokenOr:
			if len(stack) < 2 {
				return nil, fmt.Errorf("OR operator missing operands")
			}
			right := stack[len(stack)-1]
			left := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, &binaryNode{op: logicalOr, left: left, right: right})
		default:
			return nil, fmt.Errorf("unexpected token in expression")
		}
	}

	if len(stack) != 1 {
		return nil, fmt.Errorf("invalid search expression")
	}

	return stack[0], nil
}

func (n *binaryNode) eval(idx *searchDocumentIndex) bool {
	if n == nil {
		return true
	}
	switch n.op {
	case logicalAnd:
		return n.left.eval(idx) && n.right.eval(idx)
	case logicalOr:
		return n.left.eval(idx) || n.right.eval(idx)
	default:
		return false
	}
}

func (n *notNode) eval(idx *searchDocumentIndex) bool {
	if n == nil {
		return false
	}
	return !n.child.eval(idx)
}

func (n *termNode) eval(idx *searchDocumentIndex) bool {
	if n == nil || idx == nil {
		return false
	}

	if n.term == "" {
		return true
	}

	if n.isPhrase || strings.ContainsRune(n.term, ' ') {
		return wildcardMatch(idx.text, n.term)
	}

	if strings.Contains(n.term, "*") {
		for token := range idx.tokens {
			if wildcardMatch(token, n.term) {
				return true
			}
		}
		return wildcardMatch(idx.text, n.term)
	}

	if _, ok := idx.tokens[n.term]; ok {
		return true
	}

	return wildcardMatch(idx.text, n.term)
}

func wildcardMatch(value, pattern string) bool {
	if pattern == "" {
		return true
	}
	if value == "" {
		return false
	}

	if !strings.Contains(pattern, "*") {
		return strings.Contains(value, pattern)
	}

	parts := strings.Split(pattern, "*")
	index := 0

	// Handle prefix if pattern doesn't start with '*'
	if parts[0] != "" {
		if !strings.HasPrefix(value, parts[0]) {
			return false
		}
		index = len(parts[0])
	}

	for i := 1; i < len(parts); i++ {
		segment := parts[i]
		if segment == "" {
			continue
		}
		pos := strings.Index(value[index:], segment)
		if pos == -1 {
			return false
		}
		index += pos + len(segment)
	}

	if last := parts[len(parts)-1]; last != "" && !strings.HasSuffix(pattern, "*") {
		return strings.HasSuffix(value, last)
	}

	return true
}

func buildSearchDocumentIndex(job *JobRecord) *searchDocumentIndex {
	if job == nil {
		return nil
	}

	idx := &searchDocumentIndex{
		tokens: make(map[string]struct{}),
	}
	var builder strings.Builder

	addText := func(text string) {
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}
		lower := strings.ToLower(text)
		if builder.Len() > 0 {
			builder.WriteByte(' ')
		}
		builder.WriteString(lower)
		for _, token := range splitToSearchTokens(lower) {
			if token != "" {
				idx.tokens[token] = struct{}{}
			}
		}
	}

	addText(job.Title)
	addText(job.Description)
	addText(job.Engagement)
	addText(job.DurationLabel)
	addText(job.Workload)

	for _, tag := range job.Tags {
		addText(tag)
	}
	for _, skill := range job.Skills {
		addText(skill)
	}
	for _, occupation := range job.Occupations {
		addText(occupation)
	}

	if job.Category != nil {
		addText(job.Category.Name)
		addText(job.Category.Group)
		addText(job.Category.Slug)
		addText(job.Category.GroupSlug)
	}

	if job.Buyer != nil {
		addText(job.Buyer.Country)
		addText(job.Buyer.City)
	}

	idx.text = builder.String()
	return idx
}

func splitToSearchTokens(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
		switch r {
		case '_', '-', '#', '+':
			return false
		default:
			return true
		}
	})
}

func matchesSearchExpression(job *JobRecord, expr *SearchExpression) bool {
	if expr == nil || expr.root == nil {
		return true
	}
	idx := buildSearchDocumentIndex(job)
	if idx == nil {
		return false
	}
	return expr.Evaluate(idx)
}
