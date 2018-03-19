package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ABCI Response Codes
	// Base SDK reserves 0 ~ 99.
	// Coin errors reserve 100 ~ 199.
	// Lino authentication errors reserve 200 ~ 299.
	// Lino account handler errors reserve 300 ~ 399.
	// Lino post erros reserve 400 ~ 499.

	// TODO: Maybe use CodeAccountInvalidUsername
	// CodeInvalidUsername indicates the username format is invalid.
	CodeInvalidUsername sdk.CodeType = 301
	// CodeUnrecognizedUsername indicates the username is not in store.
	CodeUnrecognizedUsername sdk.CodeType = 302

	// CodePostMarshalError indicates error occurs during marshal
	CodePostMarshalError sdk.CodeType = 401
	// CodePostUnmarshalError indicates error occurs during unmarshal
	CodePostUnmarshalError sdk.CodeType = 402
	// CodePostNotFound indicates the post is not in store.
	CodePostNotFound sdk.CodeType = 403
	// CodePostCreateError occurs when create msg fails some precondition
	CodePostCreateError sdk.CodeType = 404
)
