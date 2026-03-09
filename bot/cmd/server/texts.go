package main

import "fmt"

const (
	buttonGetToken     = "Get token"
	buttonOpenCalendar = "Open calendar"
	buttonFAQ          = "FAQ"

	textAllowlisted        = "You are on the allowlist!\n\nUse MCP to call calendar functions."
	textServiceUnavailable = "Service is temporarily unavailable. Please try again later."
	textAllowlistCheckFail = "Could not check allowlist status. Please try again later."
	textAllowlistParseFail = "Allowlist check succeeded but response could not be parsed."
	textIssueTokenFail     = "Could not issue a token. Please try again later."
	textTokenParseFail     = "Token was issued but response could not be parsed."
	textTokenEmpty         = "Token response is empty. Please contact support."
	textFAQ                = "FAQ:\n\n1) What does Get token do?\nIt issues a new MCP token for your Telegram account.\n\n2) What does Open calendar do?\nIt opens the calendar page in your browser.\n\n3) I am not on the allowlist. What should I do?\nPlease contact support to get access."
)

func textNotAllowlisted(contact string) string {
	return fmt.Sprintf("You are not on the allowlist yet. Please contact %s.", contact)
}

func textTokenIssued(token string) string {
	return fmt.Sprintf("Here is your new MCP token:\n`%s`", token)
}
