package depth

import (
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

var schema = gqlparser.MustLoadSchema(
	&ast.Source{
		Name: "test.graphql",
		Input: `
		interface NameInterface {
			name: String
		}

		type Item implements NameInterface {
			scalar: String
			name: String
			list(size: Int = 10): [Item]
		}

		type ExpensiveItem implements NameInterface {
			name: String
		}

		type Named {
			name: String
		}

		union NameUnion = Item | Named

		type Query {
			scalar: String
			object: Item
			interface: NameInterface
			union: NameUnion
			customObject: Item
			list(size: Int = 10): [Item]
		}
		`,
	},
)

func requireDepth(t *testing.T, source string, limit int, exceeded bool) {
	t.Helper()
	query := gqlparser.MustLoadQuery(schema, source)

	es := &graphql.ExecutableSchemaMock{
		ComplexityFunc: func(typeName, field string, childComplexity int, args map[string]interface{}) (int, bool) {
			switch typeName + "." + field {
			case "ExpensiveItem.name":
				return 5, true
			case "Query.list", "Item.list":
				return int(args["size"].(int64)) * childComplexity, true
			case "Query.customObject":
				return 1, true
			}
			return 0, false
		},
		SchemaFunc: func() *ast.Schema {
			return schema
		},
	}

	actualExceeded := MaxDepthExceedLimit(es, query.Operations[0], limit)
	require.Equal(t, exceeded, actualExceeded)
}

func TestCalculate(t *testing.T) {
	t.Run("single scalar", func(t *testing.T) {
		const query = `
		{
			scalar
		}
		`
		requireDepth(t, query, 1, false)
	})

	t.Run("multiple scalar", func(t *testing.T) {
		const query = `
		{
			scalar
			scalar
		}
		`
		requireDepth(t, query, 1, false)
	})

	t.Run("depth 2", func(t *testing.T) {
		const query = `
		{
			list {
				scalar
			}
		}
		`
		requireDepth(t, query, 2, false)
	})

	t.Run("depth 2 exceed", func(t *testing.T) {
		const query = `
		{
			list {
				scalar
			}
		}
		`
		requireDepth(t, query, 1, true)
	})

	t.Run("more depth", func(t *testing.T) {
		const query = `
		{
			list1: list(size: 1) {
				list(size: 1) {
					list(size: 1) {
						scalar
					}
				}
			}
		}
		`
		requireDepth(t, query, 4, false)
	})

	t.Run("more depth", func(t *testing.T) {
		const query = `
		{
			list1: list(size: 1) {
				list(size: 1) {
					list(size: 1) {
						scalar
					}
				}
			}
		}
		`
		requireDepth(t, query, 4, false)
	})

	t.Run("more depth", func(t *testing.T) {
		const query = `
		{
			list1: list(size: 1) {
				list(size: 1) {
					list(size: 1) {
						scalar
					}
				}
			}
			list(size: 1) {
				list(size: 1) {
					list(size: 1) {
						scalar
					}
				}
			}
		}
		`
		requireDepth(t, query, 3, true)
	})

	t.Run("adds fragment spread", func(t *testing.T) {
		const query = `
		{
			list(size: 1) {
				...Fragment1
			}
		}

		fragment Fragment1 on Item {
			list(size: 1) {
				...Fragment2
			}
		}

		fragment Fragment2 on Item {
			list(size: 1) {
				...Fragment3
			}
		}

		fragment Fragment3 on Item {
			scalar
		}
		`
		requireDepth(t, query, 3, true)
		requireDepth(t, query, 4, false)
	})

	t.Run("adds inline fragment", func(t *testing.T) {
		const query = `
		{
			list(size: 1) {
				... on Item {
					list(size: 1) {
						... on Item {
							list(size: 1) {
								... on Item {
									scalar
								}
							}
						}
					}
				}
			}
		}
		`
		requireDepth(t, query, 3, true)
		requireDepth(t, query, 4, false)
	})

}
