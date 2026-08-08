package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/natthan/api/internal/core/cache"
	"github.com/natthan/api/internal/core/pagination"
	"github.com/natthan/api/internal/core/token"
	"github.com/natthan/api/internal/modules/person"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	Register(ctx context.Context, req RegisterRequest) (*LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	ListUsers(ctx context.Context, req ListUsersRequest) ([]User, int64, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	AddUserRole(ctx context.Context, id uuid.UUID, req AddRoleRequest) (*UserRole, error)
	RemoveUserRole(ctx context.Context, id uuid.UUID, role string) error
	ReactivateUser(ctx context.Context, id uuid.UUID) error
	DeactivateUser(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo       Repository
	personSvc  person.Service
	cache      cache.Store
	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewService(repo Repository, personSvc person.Service, cache cache.Store, jwtSecret string, accessTTL, refreshTTL time.Duration) Service {
	return &service{repo: repo, personSvc: personSvc, cache: cache, jwtSecret: jwtSecret, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "credenciais inválidas")
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "credenciais inválidas")
	}
	return s.issueTokens(ctx, user)
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*LoginResponse, error) {
	if _, err := s.repo.GetByEmail(ctx, req.Email); err == nil {
		return nil, fiber.NewError(fiber.StatusConflict, "email já cadastrado")
	}

	created, err := s.personSvc.Create(ctx, person.CreateRequest{
		FullName:      req.FullName,
		BirthDate:     req.BirthDate,
		NationalityID: req.NationalityID,
	})
	if err != nil {
		return nil, err
	}
	personID, _ := uuid.Parse(created.ID)

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("erro ao processar senha: %w", err)
	}

	user, err := s.repo.Create(ctx, User{
		ID:           uuid.New(),
		PersonID:     personID,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
	})
	if err != nil {
		return nil, err
	}

	if _, err := s.repo.AddRole(ctx, UserRole{ID: uuid.New(), UserID: user.ID, Role: RoleVisitor}); err != nil {
		return nil, err
	}

	user.Roles, _ = s.repo.GetActiveRoles(ctx, user.ID)
	return s.issueTokens(ctx, user)
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	var userID string
	if err := s.cache.Get(ctx, refreshKey(refreshToken), &userID); err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "refresh token inválido ou expirado")
	}
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "refresh token inválido")
	}
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "usuário não encontrado")
	}
	_ = s.cache.Delete(ctx, refreshKey(refreshToken), userRefreshKey(userID))
	return s.issueTokens(ctx, user)
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	var userID string
	if err := s.cache.Get(ctx, refreshKey(refreshToken), &userID); err != nil {
		return nil
	}
	return s.cache.Delete(ctx, refreshKey(refreshToken), userRefreshKey(userID))
}

func (s *service) ListUsers(ctx context.Context, req ListUsersRequest) ([]User, int64, error) {
	page := pagination.Parse(req.Page, req.Limit)
	activeOnly := req.Active != "false"
	return s.repo.List(ctx, req.Name, req.Role, activeOnly, page.Limit, page.Offset)
}

func (s *service) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "usuário não encontrado")
		}
		return nil, err
	}
	roles, err := s.repo.ListUserRoles(ctx, id)
	if err != nil {
		return nil, err
	}
	u.Roles = roles
	return u, nil
}

func (s *service) AddUserRole(ctx context.Context, id uuid.UUID, req AddRoleRequest) (*UserRole, error) {
	ur := UserRole{ID: uuid.New(), UserID: id, Role: Role(req.Role)}
	if req.ExpirationDate != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpirationDate)
		if err != nil {
			return nil, fiber.NewError(fiber.StatusBadRequest, "expiration_date inválida")
		}
		ur.ExpirationDate = &t
	}
	return s.repo.AddRole(ctx, ur)
}

func (s *service) RemoveUserRole(ctx context.Context, id uuid.UUID, role string) error {
	return s.repo.RemoveRole(ctx, id, Role(role))
}

func (s *service) ReactivateUser(ctx context.Context, id uuid.UUID) error {
	return s.repo.Reactivate(ctx, id)
}

func (s *service) DeactivateUser(ctx context.Context, id uuid.UUID) error {
	return s.repo.Deactivate(ctx, id)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func (s *service) issueTokens(ctx context.Context, user *User) (*LoginResponse, error) {
	role := effectiveRole(user.Roles)
	accessToken, err := token.Generate(user.ID.String(), user.PersonID.String(), string(role), s.jwtSecret, s.accessTTL)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar access token: %w", err)
	}
	refreshToken := uuid.New().String()
	if err := s.cache.Set(ctx, refreshKey(refreshToken), user.ID.String(), s.refreshTTL); err != nil {
		return nil, fmt.Errorf("erro ao armazenar refresh token: %w", err)
	}
	if err := s.cache.Set(ctx, userRefreshKey(user.ID.String()), refreshToken, s.refreshTTL); err != nil {
		return nil, fmt.Errorf("erro ao armazenar refresh token: %w", err)
	}
	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toUserResponse(user),
	}, nil
}

func refreshKey(t string) string       { return "refresh:" + t }
func userRefreshKey(uid string) string { return "user_refresh:" + uid }
