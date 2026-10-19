package utils

import connect "connectrpc.com/connect/v2"

func CodeToText(code connect.Code) string {
	var text string
	switch code {
	case connect.CodeAlreadyExists:
		text = "can't duplicate"
	case connect.CodeNotFound:
		text = "not found"
	default:
		text = "something went wrong"
	}
	return text
}
