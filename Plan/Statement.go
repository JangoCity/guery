package Plan

import (
	"github.com/xitongsys/guery/Config"
	"github.com/xitongsys/guery/parser"
	"github.com/xitongsys/guery/strings"
)

func NewPlanNodeFromSingleStatement(runtime *Config.ConfigRuntime, t parser.ISingleStatementContext) PlanNode {
	tt := t.(*parser.SingleStatementContext)
	return NewPlanNodeFromStatement(runtime, tt.Statement())
}

func NewPlanNodeFromStatement(runtime *Config.ConfigRuntime, t parser.IStatementContext) PlanNode {
	tt := t.(*parser.StatementContext)
	if tt.Query() != nil {
		return NewPlanNodeFromQuery(runtime, tt.Query())
	}

	//use
	if tt.USE() != nil {
		catalog, schema := runtime.Catalog, runtime.Schema

		if ct := tt.GetCatalog(); ct != nil {
			catalogNode := NewIdentifierNode(runtime, ct)
			catalog = catalogNode.GetText()
		}

		if sc := tt.GetSchema(); sc != nil {
			schemaNode := NewIdentifierNode(runtime, sc)
			schema = schemaNode.GetText()
		}

		return NewPlanUseNode(runtime, catalog, schema)
	}

	//show tables
	if tt.SHOW() != nil && tt.TABLES() != nil {
		catalog, schema := runtime.Catalog, runtime.Schema
		if qname := tt.QualifiedName(); qname != nil {
			name := NewQulifiedNameNode(runtime, qname).Result()
			names := strings.Split(name, ".")
			if len(names) == 1 {
				schema = names[0]
			} else if len(names) == 2 {
				catalog = names[0]
				schema = names[1]
			}
		}
		var like, escape *string
		return NewPlanShowTablesNode(runtime, catalog, schema, like, escape)
	}

	return nil
}
