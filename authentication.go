package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

var secretKey = os.Getenv("JWT_SECRET_KEY")

var jwtKey = []byte(secretKey)

func base64UrlEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

func base64UrlDecode(data string) ([]byte, error) {
	if l := len(data) % 4; l > 0 {
		data += strings.Repeat("=", 4-l)
	}
	return base64.URLEncoding.DecodeString(data)
}

func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func GenerateJWT(accountID, role string) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	encodedHeader := base64UrlEncode(headerBytes)

	claims := CustomClaims{
		AccountID: accountID,
		Role:      role,
		Exp:       time.Now().Add(24 * time.Hour).Unix(),
	}
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encodedClaims := base64UrlEncode(claimsBytes)

	signatureInput := encodedHeader + "." + encodedClaims
	signature := hmac.New(sha256.New, jwtKey)
	signature.Write([]byte(signatureInput))
	encodedSignature := base64UrlEncode(signature.Sum(nil))

	token := signatureInput + "." + encodedSignature
	return token, nil
}

func VerifyJWT(token string) (*CustomClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	encodedHeader, encodedClaims, encodedSignature := parts[0], parts[1], parts[2]

	signatureInput := encodedHeader + "." + encodedClaims
	expectedSignature := hmac.New(sha256.New, jwtKey)
	expectedSignature.Write([]byte(signatureInput))
	expectedSignatureBase64 := base64UrlEncode(expectedSignature.Sum(nil))

	if encodedSignature != expectedSignatureBase64 {
		return nil, fmt.Errorf("invalid token signature")
	}

	claimsBytes, err := base64UrlDecode(encodedClaims)
	if err != nil {
		return nil, fmt.Errorf("invalid claims encoding")
	}

	var claims CustomClaims
	err = json.Unmarshal(claimsBytes, &claims)
	if err != nil {
		return nil, fmt.Errorf("invalid claims format")
	}

	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token has expired")
	}

	return &claims, nil
}
