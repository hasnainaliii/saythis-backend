package usecase

import (
	"context"
	"fmt"
	"log"
	"saythis-backend/internal/src/auth/repository"
	userDomain "saythis-backend/internal/src/user/domain"
	userRepository "saythis-backend/internal/src/user/repository"
	userUseCase "saythis-backend/internal/src/user/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RegisterOrchestrator struct {
	pool     *pgxpool.Pool
	userUC   *userUseCase.UserUseCase
	authUC   *RegisterAuthUseCase
	userRepo userRepository.UserRepository
	authRepo repository.AuthRepository
	logger   *log.Logger
}

func NewRegisterOrchestrator(
	pool *pgxpool.Pool,
	userUC *userUseCase.UserUseCase,
	authUC *RegisterAuthUseCase,
	userRepo userRepository.UserRepository,
	authRepo repository.AuthRepository,
	logger *log.Logger,
) *RegisterOrchestrator {
	logger.Printf("[DEBUG] Created RegisterOrchestrator")
	return &RegisterOrchestrator{
		pool:     pool,
		userUC:   userUC,
		authUC:   authUC,
		userRepo: userRepo,
		authRepo: authRepo,
		logger:   logger,
	}
}

func (o *RegisterOrchestrator) Register(ctx context.Context, email, fullName, password string) (*userDomain.User, error) {
	o.logger.Printf("[FLOW] RegisterOrchestrator.Register triggered for Email: %s, FullName: %s", email, fullName)

	tx, err := o.pool.Begin(ctx)
	if err != nil {
		o.logger.Printf("[ERROR] [TX] Failed to begin transaction: %v", err)
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	o.logger.Printf("[DEBUG] [TX] Transaction started. Tx Address: %p", tx)

	defer func() {
		if p := recover(); p != nil {
			o.logger.Printf("[PANIC] [TX] Panic detected, rolling back: %v", p)
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	o.logger.Println("[DEBUG] [TX] Creating transaction-scoped repositories and usecases...")
	txUserRepo := o.userRepo.WithQuerier(tx)
	txAuthRepo := o.authRepo.WithQuerier(tx)

	txUserUC := o.userUC.WithRepository(txUserRepo)
	txAuthUC := o.authUC.WithRepository(txAuthRepo)

	o.logger.Println("[STEP 1] Calling UserUseCase.RegisterUser...")
	user, err := txUserUC.RegisterUser(ctx, email, fullName)
	if err != nil {
		o.logger.Printf("[ERROR] [STEP 1] User registration failed: %v. Rolling back transaction.", err)
		tx.Rollback(ctx)
		return nil, err
	}
	o.logger.Printf("[DEBUG] [STEP 1] User created successfully. ID: %s", user.ID())

	o.logger.Println("[STEP 2] Calling RegisterAuthUseCase.Execute...")
	err = txAuthUC.Execute(ctx, user.ID(), password)
	if err != nil {
		o.logger.Printf("[ERROR] [STEP 2] Auth creation failed: %v. Rolling back transaction.", err)
		tx.Rollback(ctx)
		return nil, err
	}
	o.logger.Println("[DEBUG] [STEP 2] Auth credentials created successfully.")

	o.logger.Println("[TX] Attempting to commit transaction...")
	if err := tx.Commit(ctx); err != nil {
		o.logger.Printf("[ERROR] [TX] Transaction commit failed: %v", err)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	o.logger.Println("[TX] Transaction committed successfully ✅")

	return user, nil
}
