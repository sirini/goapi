package services

import (
	"strings"
	"testing"

	"github.com/sirini/goapi/pkg/models"
)

func TestValidateBadgeDefinition(t *testing.T) {
	valid := models.AdminBadgeDefinitionParam{
		Name: "  사진전 우수상  ", Description: "  좋은 사진을 공유한 성과  ", IconKey: "trophy",
	}
	if err := validateBadgeDefinition(&valid); err != nil {
		t.Fatal(err)
	}
	if valid.Name != "사진전 우수상" || valid.Description != "좋은 사진을 공유한 성과" {
		t.Fatalf("badge fields were not normalized: %#v", valid)
	}

	for _, test := range []models.AdminBadgeDefinitionParam{
		{Name: "한", IconKey: "award"},
		{Name: "지원하지 않는 아이콘", IconKey: "custom-svg"},
		{Name: "설명이 너무 긴 업적", Description: strings.Repeat("가", 301), IconKey: "award"},
	} {
		if err := validateBadgeDefinition(&test); err == nil {
			t.Fatalf("invalid definition passed validation: %#v", test)
		}
	}
}
