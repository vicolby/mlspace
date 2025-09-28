package users

import (
	"aispace/internal/storage"
	"net/http"
)

type UserService struct {
	uow storage.UnitOfWork
}

func NewUserService(uow storage.UnitOfWork) *UserService {
	return &UserService{uow: uow}
}

func (s *UserService) GetUsers(r *http.Request) http.HandlerFunc {
	return nil
}
