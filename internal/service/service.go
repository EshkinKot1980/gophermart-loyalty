package service

type Logger interface {
	Error(message string, err error)
}
