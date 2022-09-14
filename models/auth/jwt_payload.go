package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"newTradingBot/api/security"
	"newTradingBot/models/users"
	"strings"
	"time"
)

const UserInfo = "userInfo"

type Payload struct {
	UserID 			int `json:"user_id"`
	Verified 		bool `json:"verified"`
	Expires         time.Time `json:"exp"`
}

func CreateToken(user *users.User) (*string, error) {
	payload := Payload{
		UserID: user.ID,
		Expires: time.Now().Add(tokenLifeTime),
		Verified: user.Verified,
	}

	bytes, err := json.Marshal(&payload)
	if err != nil {
		return nil, err
	}

	tokenPayload := base64.RawURLEncoding.EncodeToString(bytes)
	tokenUnHashed := fmt.Sprintf("%s%s", tokenPayload, secretKey)

	signature, err := security.GetHashedString(tokenUnHashed)
	if err != nil {
		return nil, err
	}

	token := fmt.Sprintf("%s%s%s", tokenPayload, separator, signature)

	return &token, nil
}

func VerifyToken(token string) (bool, error) {
	if token == "" {
		return false, nil
	}

	splitToken := strings.Split(token, separator)
	payload := splitToken[0]
	signature := splitToken[1]

	toVerifyString := fmt.Sprintf("%s%s", payload, secretKey)

	return security.VerifyHashedString(toVerifyString, signature)
}

func GetTokenPayload(token string) (*Payload, error) {
	splitToken := strings.Split(token, separator)
	payload := splitToken[0]

	bytes, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, err
	}

	var tp Payload
	err = json.Unmarshal(bytes, &tp)
	return &tp, err
}