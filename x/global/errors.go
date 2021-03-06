package global

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lino-network/lino/types"
)

// ErrRegisterExpiredEvent - error when register event time is before current timestamp
func ErrRegisterExpiredEvent(unixTime int64) sdk.Error {
	return types.NewError(types.CodeRegisterExpiredEvent, fmt.Sprintf("register event at expired time %v", unixTime))
}

// ErrAmountOfConsumptionExponent - error when transfer global parameter from rat to float failed
func ErrAmountOfConsumptionExponent() sdk.Error {
	return types.NewError(types.CodeFailedToGetAmountOfConsumptionExponent, "global parameter error")
}

// ErrGetPastDay - error if get past day is negative
func ErrGetPastDay() sdk.Error {
	return types.NewError(types.CodeFailedToGetAmountOfConsumptionExponent, "get past day failed")
}
