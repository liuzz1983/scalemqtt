package main

type Auth interface {
	Auth(userName string, password string) bool
}

type NullAuth struct{}

func (auth *NullAuth) Auth(userName string, password string) bool {
	return true
}
