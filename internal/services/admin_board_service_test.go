package services

import (
	"strings"
	"testing"

	"github.com/sirini/goapi/pkg/models"
)

func TestValidateBoardModifyTargetRejectsMissingGroup(t *testing.T) {
	param := models.AdminBoardModifyParam{
		AdminBoardCreateParam: models.AdminBoardCreateParam{GroupUid: 0},
		BoardUid:              23,
	}

	err := validateBoardModifyTarget(param, 23, false)
	if err == nil || !strings.Contains(err.Error(), "group") {
		t.Fatalf("missing group error = %v", err)
	}
}

func TestValidateBoardModifyTargetRejectsMismatchedBoard(t *testing.T) {
	param := models.AdminBoardModifyParam{
		AdminBoardCreateParam: models.AdminBoardCreateParam{GroupUid: 17},
		BoardUid:              23,
	}

	err := validateBoardModifyTarget(param, 24, true)
	if err == nil || !strings.Contains(err.Error(), "board") {
		t.Fatalf("mismatched board error = %v", err)
	}
}

func TestValidateBoardModifyTargetAcceptsExistingTargets(t *testing.T) {
	param := models.AdminBoardModifyParam{
		AdminBoardCreateParam: models.AdminBoardCreateParam{GroupUid: 17},
		BoardUid:              23,
	}

	if err := validateBoardModifyTarget(param, 23, true); err != nil {
		t.Fatalf("valid target error = %v", err)
	}
}
