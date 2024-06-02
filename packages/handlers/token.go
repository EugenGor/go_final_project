package handlers

import (
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/golang-jwt/jwt/v5"
)

// создание токена
func createToken(userPassedPassword string, encKey string) (string, error) {
	secret := []byte(encKey)

	signedPassword := HashPassword([]byte(userPassedPassword), secret) //хэшированый пароль и секретное слово

	// создание payload
	//добавляется хэшированный пароль (signedPassword) в качестве значения с ключом "password"
	claims := jwt.MapClaims{
		"password": signedPassword,
	}

	// создается новый JWT токен (jwtToken) с указанием метода подписи HS256
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// выполняется подпись JWT токена с использованием заданного секретного ключа (secret)
	signedToken, err := jwtToken.SignedString(secret)
	if err != nil {
		log.Printf("failed to sign jwt: %s\n", err)
		return "", err
	}

	fmt.Println("Result token: " + string(signedToken[:]))

	//возвращается подписанный токен в виде строки
	return signedToken, nil
}

// функция принимает срез байтов `password` и `secretKey`, хеширует их с использованием алгоритма SHA-256, суммирует результаты и возвращает хеш в виде строки
func HashPassword(password []byte, secretKey []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(append(password, secretKey...)))
}
