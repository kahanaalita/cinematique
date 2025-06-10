package handlers

// AuthService Определяет интерфейс для операций аутентификации
type AuthService interface {
	// Register Создает нового пользователя с данными учетными данными
	Register(username, password, role string) (int, error)
	// Login Аутентифицирует пользователя и возвращает токен JWT
	Login(username, password string) (string, error)
}
