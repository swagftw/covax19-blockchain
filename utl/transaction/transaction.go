package transaction

import "context"

type Transaction interface {
	Run(context.Context, func(context.Context) error) error
}
