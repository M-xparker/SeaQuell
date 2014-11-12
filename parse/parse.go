package parse

import (
	"strconv"
)

type Tree struct {
	lex  *lexer
	Root interface{}
}

func New(input string) *Tree {
	t := new(Tree)
	t.lex = lex(input)
	return t
}

func (t *Tree) nextToken() token {
	return t.lex.nextItem()
}

func (t *Tree) Parse() {
	for {
		i := t.nextToken()
		if i.typ == tokenEOF {
			break
		}
		n := t.match(i)
		if t.Root == nil {
			t.Root = n
		}
	}
}

func (t *Tree) match(i token) interface{} {
	switch i.typ {
	case tokenSelect:
		return t.selectStatement()
	case tokenCreate:
		return t.create()
	case tokenInsert:
		return t.insert()
	}
	return nil
}

func (t *Tree) insert() interface{} {
	for {
		switch i := t.nextToken(); i.typ {
		case tokenInto:
			return t.insertInto()
		}
	}
	return nil
}

func (t *Tree) insertInto() interface{} {
	node := &NodeInsertInto{}
Loop:
	for {
		switch i := t.nextToken(); i.typ {
		case tokenIdentifier:
			node.table = i.val
		case tokenNumber:
			var number interface{}
			var err error
			n, err := strconv.Atoi(i.val)
			number = int64(n)
			if err != nil {
				n2, err := strconv.ParseFloat(i.val, 64)
				number = n2
				if err != nil {
					break Loop
				}
			}
			node.values = append(node.values, number)
		case tokenSemiColon:
			break Loop
		}
	}
	return node
}

func (t *Tree) create() interface{} {
	for {
		switch i := t.nextToken(); i.typ {
		case tokenTable:
			return t.createTable()
		}
	}
	return nil
}

func (t *Tree) createTable() interface{} {
	node := &NodeCreateTable{}
Loop:
	for {
		switch i := t.nextToken(); i.typ {
		case tokenIdentifier:
			if node.table == "" {
				node.table = i.val
				break
			}
			node.columns = append(node.columns, &Column{name: i.val})
		case tokenDataType:
			col := node.columns[len(node.columns)-1]
			col.typ = i.val
		case tokenSemiColon:
			break Loop
		}
	}
	return node
}

func (t *Tree) selectStatement() *NodeSelectFrom {
	s := new(NodeSelectFrom)
	var hasFrom bool
Loop:
	for {
		switch i := t.nextToken(); i.typ {
		case tokenIdentifier:
			if hasFrom {
				s.table = i.val
				break Loop
			}
			s.fields = append(s.fields, &FieldNode{name: i.val})
		case tokenFrom:
			hasFrom = true
		case tokenStar:
			s.hasStar = true
		case tokenSpace:
		case tokenSemiColon:
			break Loop
		default:
			panic("unexpected token" + i.String())
		}
	}
	return s
}
