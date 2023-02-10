package depth

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
)

func MaxDepthExceedLimit(es graphql.ExecutableSchema, op *ast.OperationDefinition, limit int) bool {
	walker := walker{
		es:     es,
		schema: es.Schema(),
	}
	_, exceeded := walker.selectionSetDepth(op.SelectionSet, false, limit)
	return exceeded
}

type walker struct {
	es     graphql.ExecutableSchema
	schema *ast.Schema
}

func (g walker) selectionSetDepth(set ast.SelectionSet, isFragment bool, limit int) (int, bool) {
	maxChildrenDepth := 0
	for _, selection := range set {
		depth := g.walk(selection, limit)
		if depth > maxChildrenDepth {
			maxChildrenDepth = depth
		}
		if maxChildrenDepth > limit {
			return maxChildrenDepth, true
		}
	}
	if !isFragment {
		// if fragment, we should not count fragment statement itself as another level of depth.
		maxChildrenDepth += 1
	}
	return maxChildrenDepth, false
}

// dfs
func (g walker) walk(selection ast.Selection, limit int) int {
	var depth int
	switch s := selection.(type) {
	case *ast.Field:
		fieldDefinition := g.schema.Types[s.Definition.Type.Name()]
		// Ignore IntrospectionQuery (query from graphgl playground)
		if fieldDefinition.Name == "__Schema" {
			return 1
		}
		depth, _ = g.selectionSetDepth(s.SelectionSet, false, limit)
	case *ast.FragmentSpread:
		depth, _ = g.selectionSetDepth(s.Definition.SelectionSet, true, limit)
	case *ast.InlineFragment:
		depth, _ = g.selectionSetDepth(s.SelectionSet, true, limit)
	}
	return depth
}
