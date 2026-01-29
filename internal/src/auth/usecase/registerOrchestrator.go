package usecase

import (
	"context"
	"fmt"
	"saythis-backend/internal/src/auth/repository"
	userDomain "saythis-backend/internal/src/user/domain"
	userRepository "saythis-backend/internal/src/user/repository"
	userUseCase "saythis-backend/internal/src/user/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type RegisterOrchestrator struct {
	pool     *pgxpool.Pool
	userUC   *userUseCase.UserUseCase
	authUC   *RegisterAuthUseCase
	userRepo userRepository.UserRepository
	authRepo repository.AuthRepository
}

func NewRegisterOrchestrator(
	pool *pgxpool.Pool,
	userUC *userUseCase.UserUseCase,
	authUC *RegisterAuthUseCase,
	userRepo userRepository.UserRepository,
	authRepo repository.AuthRepository,

) *RegisterOrchestrator {
	return &RegisterOrchestrator{
		pool:     pool,
		userUC:   userUC,
		authUC:   authUC,
		userRepo: userRepo,
		authRepo: authRepo,
	}
}

func (o *RegisterOrchestrator) Register(ctx context.Context, email, fullName, password string) (*userDomain.User, error) {

	tx, err := o.pool.Begin(ctx)
	if err != nil {
		zap.S().Errorw("Failed to begin transaction", "email", email, "error", err)
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				zap.S().Errorw("Failed to rollback on panic", "error", rbErr)
			}
			panic(p)
		}
	}()

	txUserRepo := o.userRepo.WithQuerier(tx)
	txAuthRepo := o.authRepo.WithQuerier(tx)

	txUserUC := o.userUC.WithRepository(txUserRepo)
	txAuthUC := o.authUC.WithRepository(txAuthRepo)

	user, err := txUserUC.CreateUser(ctx, email, fullName)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			zap.S().Errorw("Failed to rollback transaction", "error", rbErr)
		}
		return nil, err
	}

	err = txAuthUC.Execute(ctx, user.ID(), password)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			zap.S().Errorw("Failed to rollback transaction", "error", rbErr)
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		zap.S().Errorw("Failed to commit transaction", "email", email, "error", err)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	zap.S().Debugw("Registration transaction completed", "email", email, "user_id", user.ID())
	return user, nil
}
