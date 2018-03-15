package auth

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lino-network/lino/types"
)

type CodeType = sdk.CodeType

// NOTE: Don't stringer this, we'll put better messages in later.
func codeToDefaultMsg(code CodeType) string {
	switch code {
	case types.CodeInvalidUsername:
		return "Invalid username"
	case types.CodeUnrecognizedUsername:
		return "Unrecognized username"
	default:
		return sdk.CodeToDefaultMsg(code)
	}
}

func ErrUnrecognizedUsername(accKey types.AccountKey) sdk.Error {
	return newError(types.CodeUnrecognizedUsername, fmt.Sprintf("Unrecognized username: %s", accKey))
}

func msgOrDefaultMsg(msg string, code CodeType) string {
	if msg != "" {
		return msg
	} else {
		return codeToDefaultMsg(code)
	}
}

func newError(code CodeType, msg string) sdk.Error {
	msg = msgOrDefaultMsg(msg, code)
	return sdk.NewError(code, msg)
}
