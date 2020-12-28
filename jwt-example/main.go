package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AtExpires    int64
	RtExpires    int64
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var (
	userOne = User{
		ID:       "1",
		Username: "username",
		Password: "password",
	}

	redisClient *redis.Client

	expiredTokenMessage = "Token is expired"
)

func IsUserAble(userId string, token *jwt.Token) bool {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	if claims["user_id"] != userId {
		return false
	}

	return true
}

// VerifyToken is used to verify a JWT Token
func VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Make sure token is verified against method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("method to verify failed")
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	return token, err
}

func ExtractToken(r *http.Request) string {
	bearerString := r.Header.Get("Authorization")

	bearerSplit := strings.Split(bearerString, " ")
	if len(bearerSplit) == 2 {
		return bearerSplit[1]
	}

	return ""
}

func CreateToken(id string) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Second * 60).Unix()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 60).Unix()

	atClaims := jwt.MapClaims{}
	atClaims["user_id"] = id
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)

	var err error

	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}

	rtClaims := jwt.MapClaims{}
	rtClaims["user_id"] = id
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		return nil, err
	}

	return td, nil
}

func CreateRefreshEntry(id string, td *TokenDetails) error {
	err := redisClient.Set(id, td.RefreshToken, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func Login(c *gin.Context) {
	var u User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, "Invalid request")
		return
	}

	if userOne.Username != u.Username || userOne.Password != u.Password {
		c.JSON(http.StatusUnauthorized, "Please provide valid login details")
		return
	}

	td, err := CreateToken(userOne.ID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}

	err = CreateRefreshEntry(userOne.ID, td)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	tokens := map[string]string{
		"access_token":  td.AccessToken,
		"refresh_token": td.RefreshToken,
	}

	c.JSON(http.StatusOK, tokens)
	return
}

func initializeRedis() {
	dsn := os.Getenv("REDIS_DSN")
	if dsn == "" {
		dsn = "localhost:6379"
	}
	redisClient = redis.NewClient(&redis.Options{
		Addr: dsn,
	})
}

func CreateTodo(c *gin.Context) {
	userId := c.Param("userId")

	token, err := VerifyToken(c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	isAble := IsUserAble(userId, token)
	if !isAble {
		c.JSON(http.StatusBadRequest, "unauthenticated")
		return
	}

	c.JSON(http.StatusOK, "You are authenticated")
	return

}

func ValidateRefreshToken(refreshToken string) (*jwt.Token, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Method to verify failed")
		}
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})

	// Only fail if reason is not the refresh token has expired
	if err != nil {
		return nil, err
	}

	return token, nil
}

func GenerateNewTokens(userId string, token *jwt.Token) (map[string]string, error) {
	able := IsUserAble(userId, token)
	if !able {
		return nil, fmt.Errorf("unauthorized")
	}

	// Verify that the refresh token has not expired then create new ones
	ts, err := CreateToken(userId)
	if err != nil {
		return nil, fmt.Errorf("error creating jwts: %s", err.Error())
	}

	tokens := map[string]string{
		"access_token":  ts.AccessToken,
		"refresh_token": ts.RefreshToken,
	}

	err = CreateRefreshEntry(userId, ts)
	if err != nil {
		return nil, fmt.Errorf("error creating entry in redis: %s", err.Error())
	}

	return tokens, nil

}

func RefreshToken(c *gin.Context) {
	// Should check if refresh token and user id was sent in payload
	// If they weren't provided, this handler should fail
	type RefreshPayload struct {
		UserId       string `json:"userId"`
		RefreshToken string `json:"refreshToken"`
	}

	var rPayload RefreshPayload

	if err := c.ShouldBindJSON(&rPayload); err != nil {
		c.JSON(http.StatusBadRequest, "refresh token is invalid")
		return
	}

	token, err := ValidateRefreshToken(rPayload.RefreshToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, "refresh token can not be validated")
		return
	}

	tokens, err := GenerateNewTokens(rPayload.UserId, token)
	if err != nil {
		c.JSON(http.StatusBadRequest, "new tokens could not be created")
		return
	}

	c.JSON(http.StatusOK, tokens)
	return
}

func InvalidateKeyInRedis(userId string) error {
	err := redisClient.Del(userId).Err()
	if err != nil {
		return err
	}

	return nil
}

func Logout(c *gin.Context) {
	// Primary function of this handler is to purge key in Redis
	var body map[string]interface{}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, "Can not unmarshal body")
		return
	}

	err := InvalidateKeyInRedis(body["userId"].(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Error deleting key")
		return
	}

	c.JSON(http.StatusOK, "User successfully logged out")
	return
}

// For Access tokens not only

// Refresh tokens should be stored in Redis with some unique id per user
// Once the user logs out, we should invalidate, purge that key in Redis
// The access token will still be valid until it expires
// But once we have no refresh token in Redis, we should not be able to generate new pairs of access and refresh tokens
func main() {
	initializeRedis()

	r := gin.Default()

	r.POST("/login", Login)
	r.PUT("/todo/:userId", CreateTodo)
	r.POST("/token/refresh", RefreshToken)
	r.POST("/logout", Logout)

	log.Fatal(r.Run(":5000"))
}
