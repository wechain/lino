package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lino-network/lino/types"
)

type contextKey int

const (
	contextKeySigners contextKey = iota
)

// WithSigners creates a new context with accounts
func WithSigners(ctx sdk.Context, accounts []types.Account) sdk.Context {
	return ctx.WithValue(contextKeySigners, accounts)
}
