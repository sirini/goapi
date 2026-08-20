package services

import (
	"fmt"
	"strings"

	"github.com/sirini/goapi/internal/repositories"
	"github.com/sirini/goapi/pkg/models"
)

type PushService interface {
	RegisterDevice(userUid uint, param models.PushDeviceParam) error
	UnregisterDevice(userUid uint, param models.PushDeviceParam) error
}

type NuboPushService struct {
	repo repositories.PushRepository
}

func NewNuboPushService(repo repositories.PushRepository) *NuboPushService {
	return &NuboPushService{repo: repo}
}

func (s *NuboPushService) RegisterDevice(userUid uint, param models.PushDeviceParam) error {
	param.Token = strings.TrimSpace(param.Token)
	param.Platform = strings.ToLower(strings.TrimSpace(param.Platform))
	if userUid < 1 || len(param.Token) < 20 || len(param.Token) > 512 || param.Platform != "android" {
		return fmt.Errorf("invalid push device")
	}
	return s.repo.SaveDevice(userUid, param.Token, param.Platform)
}

func (s *NuboPushService) UnregisterDevice(userUid uint, param models.PushDeviceParam) error {
	param.Token = strings.TrimSpace(param.Token)
	if userUid < 1 || param.Token == "" {
		return fmt.Errorf("invalid push device")
	}
	return s.repo.RemoveDevice(userUid, param.Token)
}
