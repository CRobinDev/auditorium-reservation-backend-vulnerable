package validator

import "github.com/bytedance/sonic"

type validationError struct {
	Tag         string `json:"tag"`
	Param       string `json:"param"`
	Translation string `json:"translation"`
}

type ValidationErrorsResponse []map[string]validationError

func (v ValidationErrorsResponse) Error() string {
	j, err := sonic.Marshal(v)
	if err != nil {
		return ""
	}

	return string(j)
}

func (v ValidationErrorsResponse) Serialize() any {
	return v
}
