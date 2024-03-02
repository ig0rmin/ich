package user

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	*Repository
}

func NewService(r *Repository) *Service {
	return &Service{r}
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedPassword), nil
}

func checkPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (s *Service) CreateUser(ctx context.Context, req *CreateUserReq) (*CreateUserRes, error) {
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	u := &User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}

	r, err := s.Repository.CreateUser(ctx, u)
	if err != nil {
		return nil, err
	}

	res := &CreateUserRes{
		ID:       strconv.Itoa(int(r.ID)),
		Username: r.Username,
		Email:    r.Email,
	}

	return res, nil
}

type JWTClaims struct {
	ID       string `json:"id"`
	UserName string `json:"username"`
	jwt.RegisteredClaims
}

func (s *Service) Login(ctx context.Context, req *LoginUserReq) (*LoginUserRes, error) {
	user, err := s.Repository.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if err := checkPassword(req.Password, user.Password); err != nil {
		return nil, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
		ID:       strconv.Itoa(user.ID),
		UserName: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    strconv.Itoa(user.ID),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	})

	const secretKey = "afiu/j8piuN54s5l8GoWEQ=="

	ss, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return &LoginUserRes{}, err
	}

	return &LoginUserRes{JWT: ss, Username: user.Username, ID: strconv.Itoa(user.ID)}, nil
}
