package main

import "fmt"

const (
	buttonGetToken     = "Получить токен"
	buttonOpenCalendar = "Открыть календарь"
	buttonFAQ          = "FAQ"

	textAllowlisted        = "Вы в списке доступа!\n\nИспользуйте MCP, чтобы вызывать функции календаря."
	textServiceUnavailable = "Сервис временно недоступен. Пожалуйста, попробуйте позже."
	textAllowlistCheckFail = "Не удалось проверить доступ. Пожалуйста, попробуйте позже."
	textAllowlistParseFail = "Проверка доступа выполнена, но не удалось разобрать ответ сервера."
	textIssueTokenFail     = "Не удалось выдать токен. Пожалуйста, попробуйте позже."
	textTokenParseFail     = "Токен выдан, но не удалось разобрать ответ сервера."
	textTokenEmpty         = "Ответ с токеном пустой. Пожалуйста, обратитесь в поддержку."
	textFAQ                = "FAQ:\n\n1) Что делает кнопка «Получить токен»?\nОна выдает новый MCP-токен для вашего Telegram-аккаунта.\n\n2) Что делает кнопка «Открыть календарь»?\nОна открывает страницу календаря в вашем браузере.\n\n3) Меня нет в списке доступа. Что делать?\nПожалуйста, обратитесь в поддержку, чтобы получить доступ."
)

func textNotAllowlisted(contact string) string {
	return fmt.Sprintf("Вас пока нет в списке доступа. Пожалуйста, свяжитесь с %s.", contact)
}

func textTokenIssued(token string) string {
	return fmt.Sprintf("Ваш новый MCP-токен:\n`%s`", token)
}
