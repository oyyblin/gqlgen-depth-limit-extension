package depth

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/errcode"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const errDepthLimit = "DEPTH_LIMIT_EXCEEDED"

// DepthLimit allows you to define a limit on query depth
//
// If a query is submitted that exceeds the limit, a 422 status code will be returned.
type DepthLimit struct {
	Func func(ctx context.Context, rc *graphql.OperationContext) int

	es graphql.ExecutableSchema
}

var _ interface {
	graphql.OperationContextMutator
	graphql.HandlerExtension
} = &DepthLimit{}

const depthExtension = "DepthLimit"

// FixedDepthLimit sets a depth limit that does not change
func FixedDepthLimit(limit int) *DepthLimit {
	return &DepthLimit{
		Func: func(ctx context.Context, rc *graphql.OperationContext) int {
			return limit
		},
	}
}

func (e DepthLimit) ExtensionName() string {
	return depthExtension
}

func (e *DepthLimit) Validate(schema graphql.ExecutableSchema) error {
	if e.Func == nil {
		return fmt.Errorf("DepthLimit func can not be nil")
	}
	e.es = schema
	return nil
}

func (e DepthLimit) MutateOperationContext(ctx context.Context, rc *graphql.OperationContext) *gqlerror.Error {
	op := rc.Doc.Operations.ForName(rc.OperationName)
	limit := e.Func(ctx, rc)

	if MaxDepthExceedLimit(e.es, op, limit) {
		err := gqlerror.Errorf("operation exceeds the depth limit of %d", limit)
		errcode.Set(err, errDepthLimit)
		return err
	}

	return nil
}
