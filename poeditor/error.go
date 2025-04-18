package poeditor

import (
	"fmt"
	"strconv"
)

const RateLimitErrorCode = 4048

// Error represents a known error from POEditor API.
//
// See: https://poeditor.com/docs/error_codes
type Error struct {
	Code        int
	Message     string
	Description string
}

func TryNewErrorFromResponse(resp response) error {
	code, _ := strconv.Atoi(resp.Code)

	if code == 200 {
		return nil
	}

	return &Error{
		Code:        code,
		Message:     resp.Message,
		Description: errorDescriptionFromCode(code),
	}
}

func errorDescriptionFromCode(code int) string {
	switch code {
	case 401:
		return "Every API request requires an API token."
	case 4011:
		return "There is no account associated with the API token used for this request."
	case 4012:
		return "Please use a POST request (only POST requests accepted)."
	case 403:
		return "The resource you are trying to access is restricted for this account. (wrong project ID)"
	case 4030:
		return "The endpoint requires a token with writing permissions."
	case 4031:
		return "API Token is valid but the account doesnt't have API access."
	case 4032:
		return "The account reached the maximum number of strings."
	case 4033:
		return "There's an import in progress."
	case 4034:
		return "The project has been archived and cannot be accessed unless restored."
	case 404:
		return "The method used is not supported."
	case 4040:
		// custom error message
		return ""
	case 4042:
		return "Parameter -data- must be a JSON object"
	case 4043:
		return "This language is not in the list of available languages."
	case 4044:
		return "The language code does not correspond to any of the languages in this project."
	case 4045:
		return "Parameter -language- is missing or empty."
	case 4046:
		return "The file could not be parsed."
	case 4047:
		return "Wrong export file format chosen."
	case RateLimitErrorCode:
		return "File uploads are limited according to plan."
	case 4049:
		return "The parameter –updating- could not be found in the request."
	case 4050:
		return "The language you are trying to add already exists in the project."
	case 4051:
		return "The download link (export file) coult not be found on server."
	case 4052:
		return "Download URLs expire after 10 minutes."
	case 4053:
		return "The URL is valid but the project/language has been deleted."
	case 429:
		return "Too many pending requests in your queue (over 200)."
	default:
		return ""
	}
}

func (e Error) Error() string {
	msg := fmt.Sprintf("POEditor API returned %d error code - %s", e.Code, e.Message)

	if e.Description != "" {
		msg += ": " + e.Description
	}

	return msg
}
