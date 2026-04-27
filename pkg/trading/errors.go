package trading

import (
	"fmt"
	"strings"
)

type Errors []responseError

func (err Errors) Error() string {
	var errs []string

	for _, e := range err {
		errs = append(errs, fmt.Sprintf("%#v", e))
	}

	return strings.Join(errs, ", ")
}

func (errs Errors) RevisionError() bool {
	for _, err := range errs {
		if err.ErrorCode == 10039 || err.ErrorCode == 10029 || err.ErrorCode == 21916916 || err.ErrorCode == 21916923 || err.ErrorCode == 21919028 {
			return true
		}
	}

	return false
}

func (errs Errors) ListingEnded() bool {
	for _, err := range errs {
		if err.ErrorCode == 291 || err.ErrorCode == 240 {
			return true
		}
	}

	return false
}

func (errs Errors) InvalidAuthToken() bool {
	for _, err := range errs {
		if err.ErrorCode == 931 {
			return true
		}
	}

	return false
}

func (errs Errors) ListingDeleted() bool {
	for _, err := range errs {
		if err.ErrorCode == 17 {
			return true
		}
	}

	return false
}

type httpError struct {
	statusCode int
	body       []byte
}

func (err httpError) Error() string {
	return fmt.Sprintf("%d - %s", err.statusCode, err.body)
}
