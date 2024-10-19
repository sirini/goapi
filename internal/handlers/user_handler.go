package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/sirini/goapi/internal/services"
	"github.com/sirini/goapi/pkg/models"
	"github.com/sirini/goapi/pkg/utils"
)

// 응답 구조체
type ResponseUserInfo struct {
	Success bool             `json:"success"`
	Error   string           `json:"error"`
	Result  *models.UserInfo `json:"result"`
}

// 사용자 정보 열람
func LoadUserInfoHandler(s *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := ResponseUserInfo{
			Success: false,
			Error:   "",
		}

		targetUserUidStr := r.URL.Query().Get("targetUserUid")
		if targetUserUidStr == "" {
			response.Error = "Missing targetUserUid parameter"
			utils.ResponseJSON(w, response)
			return
		}

		targetUserUid, err := strconv.ParseUint(targetUserUidStr, 10, 32)
		if err != nil {
			response.Error = "Invalid targetUserUid parameter"
			utils.ResponseJSON(w, response)
			return
		}

		userInfo, err := s.UserService.GetUserInfo(uint(targetUserUid))
		if err != nil {
			response.Error = strings.Join([]string{"User not found, given: ", targetUserUidStr}, "")
			utils.ResponseJSON(w, response)
			return
		}

		response.Result = userInfo
		utils.ResponseJSON(w, response)
	}
}
