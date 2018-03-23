package validator

import (
	acc "github.com/lino-network/lino/tx/account"
)

// Validator Account
type ValidatorAccount struct {
	validatorName acc.AccountKey `json:"validator_name"`
	votes         []Vote         `json:"votes"`
	totalWeight   int64          `json:"total_weight"`
	deposit       int64          `json:"deposit"`
}

// Validator candidate list
type ValidatorList struct {
	validatorListKey acc.AccountKey     `json:"validator_list_key"`
	validators       []ValidatorAccount `json:"validators"`
}

// User's vote
type Vote struct {
	voter         acc.AccountKey `json:"voter"`
	weight        int64          `json:"weight"`
	validatorName acc.AccountKey `json:"validator_name"`
}
