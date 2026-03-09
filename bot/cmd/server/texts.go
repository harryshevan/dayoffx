package main

import "fmt"

const (
	buttonGetToken     = "Получить токен"
	buttonReissueToken = "Перевыпустить токен"
	buttonOpenCalendar = "Открыть календарь"

	textAllowlisted        = "Ты с нами!\n\nИспользуй MCP, чтобы вызывать функции календаря."
	textServiceUnavailable = "Сервис временно недоступен"
	textAllowlistCheckFail = "Не удалось проверить доступ"
	textAllowlistParseFail = "Не удалось выполнить проверку доступа"
	textIssueTokenFail     = "Не удалось выдать токен"
	textTokenParseFail     = "Ошибка при выдаче токена"
	textTokenEmpty         = "Не удалось выдать токен"
)

func textNotAllowlisted(contact string) string {
	return fmt.Sprintf("Вас пока нет в списке доступа. Пожалуйста, свяжитесь с %s.", contact)
}

func textTokenIssued(token string) string {
	return fmt.Sprintf("Ваш новый MCP-токен:\n`%s`", token)
}
