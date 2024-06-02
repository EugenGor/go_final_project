package config

import (
	"fmt"
)

/*
Структура Config:
AppPassword - пароль приложения
EncryptionSecretKey - секретный ключ для шифрования
ApiPort - порт API
*/
type Config struct {
	AppPassword         string
	EncryptionSecretKey string
	ApiPort             string
}

// NewConfig конструктор объекта конфигурации приложения
func NewConfig(appPass string, encKey string, apiPort string) (*Config, error) {
	//если appPass и encKey пустые, то вызывается ошибка "invalid config"
	if appPass == "" || encKey == "" {
		return nil, fmt.Errorf("invalid config")
	}
	//если apiPort пустой, устанавливает значение "7540" по умолчанию
	if apiPort == "" {
		apiPort = "7540"
	}
	return &Config{AppPassword: appPass, EncryptionSecretKey: encKey, ApiPort: apiPort}, nil
}
