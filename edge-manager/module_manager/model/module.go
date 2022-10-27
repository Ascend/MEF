package model

type Module interface {
	Name() string
	Start()
	Enable() bool
}
