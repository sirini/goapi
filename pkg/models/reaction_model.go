package models

import "fmt"

// DB reaction_type 코드와 API 문자열 식별자를 고정한다.
const (
	REACTION_NONE      ReactionType = 0
	REACTION_LIKE      ReactionType = 1
	REACTION_BEST      ReactionType = 2
	REACTION_FACEPALM  ReactionType = 3
	REACTION_HMM       ReactionType = 4
	REACTION_LAUGH     ReactionType = 5
	REACTION_CELEBRATE ReactionType = 6
	REACTION_FIRE      ReactionType = 7
	REACTION_SUPPORT   ReactionType = 8
	REACTION_SAD       ReactionType = 9
	REACTION_EYES      ReactionType = 10
)

type ReactionType uint8

type Reaction = string

// 취소는 JSON null로 구분하며, 빈 문자열은 거부한다.
type ReactionParam struct {
	Reaction *Reaction `json:"reaction"`
}

// PATCH /board/reaction 본문. reaction은 네 종류 문자열이나 명시적 null만 허용한다.
type BoardReactionBody struct {
	BoardUid uint      `json:"boardUid"`
	PostUid  uint      `json:"postUid"`
	Reaction *Reaction `json:"reaction"`
}

// PATCH /comment/reaction 본문. reaction은 네 종류 문자열이나 명시적 null만 허용한다.
type CommentReactionBody struct {
	BoardUid   uint      `json:"boardUid"`
	CommentUid uint      `json:"commentUid"`
	Reaction   *Reaction `json:"reaction"`
}

var REACTIONS = struct {
	LIKE      Reaction
	BEST      Reaction
	FACEPALM  Reaction
	HMM       Reaction
	LAUGH     Reaction
	CELEBRATE Reaction
	FIRE      Reaction
	SUPPORT   Reaction
	SAD       Reaction
	EYES      Reaction
}{LIKE: "like", BEST: "best", FACEPALM: "facepalm", HMM: "hmm", LAUGH: "laugh", CELEBRATE: "celebrate", FIRE: "fire", SUPPORT: "support", SAD: "sad", EYES: "eyes"}

func (v ReactionType) APIValue() Reaction {
	switch v {
	case REACTION_LIKE:
		return REACTIONS.LIKE
	case REACTION_BEST:
		return REACTIONS.BEST
	case REACTION_FACEPALM:
		return REACTIONS.FACEPALM
	case REACTION_HMM:
		return REACTIONS.HMM
	case REACTION_LAUGH:
		return REACTIONS.LAUGH
	case REACTION_CELEBRATE:
		return REACTIONS.CELEBRATE
	case REACTION_FIRE:
		return REACTIONS.FIRE
	case REACTION_SUPPORT:
		return REACTIONS.SUPPORT
	case REACTION_SAD:
		return REACTIONS.SAD
	case REACTION_EYES:
		return REACTIONS.EYES
	default:
		return ""
	}
}

func ParseReaction(v Reaction) (ReactionType, error) {
	switch v {
	case REACTIONS.LIKE:
		return REACTION_LIKE, nil
	case REACTIONS.BEST:
		return REACTION_BEST, nil
	case REACTIONS.FACEPALM:
		return REACTION_FACEPALM, nil
	case REACTIONS.HMM:
		return REACTION_HMM, nil
	case REACTIONS.LAUGH:
		return REACTION_LAUGH, nil
	case REACTIONS.CELEBRATE:
		return REACTION_CELEBRATE, nil
	case REACTIONS.FIRE:
		return REACTION_FIRE, nil
	case REACTIONS.SUPPORT:
		return REACTION_SUPPORT, nil
	case REACTIONS.SAD:
		return REACTION_SAD, nil
	case REACTIONS.EYES:
		return REACTION_EYES, nil
	default:
		return REACTION_NONE, fmt.Errorf("unknown reaction: %q", v)
	}
}

// 코드 오름차순 종류 목록. 집계 초기화와 DTO 변환에 사용한다.
var ReactionCodeList = []ReactionType{
	REACTION_LIKE, REACTION_BEST, REACTION_FACEPALM, REACTION_HMM,
	REACTION_LAUGH, REACTION_CELEBRATE, REACTION_FIRE, REACTION_SUPPORT, REACTION_SAD, REACTION_EYES,
}

func NewReactionCounts() ReactionCounts {
	counts := make(ReactionCounts)
	for _, code := range ReactionCodeList {
		counts[code] = 0
	}
	return counts
}

func (counts ReactionCounts) DTO() ReactionCountsDTO {
	return ReactionCountsDTO{
		Like:      counts[REACTION_LIKE],
		Best:      counts[REACTION_BEST],
		Facepalm:  counts[REACTION_FACEPALM],
		Hmm:       counts[REACTION_HMM],
		Laugh:     counts[REACTION_LAUGH],
		Celebrate: counts[REACTION_CELEBRATE],
		Fire:      counts[REACTION_FIRE],
		Support:   counts[REACTION_SUPPORT],
		Sad:       counts[REACTION_SAD],
		Eyes:      counts[REACTION_EYES],
	}
}

// 종류별 집계는 0이어도 항상 응답에 포함한다.
type ReactionCounts map[ReactionType]uint

// JSON 계약은 종류 문자열을 키로 가진 고정 순서 DTO를 따른다.
type ReactionCountsDTO struct {
	Like      uint `json:"like"`
	Best      uint `json:"best"`
	Facepalm  uint `json:"facepalm"`
	Hmm       uint `json:"hmm"`
	Laugh     uint `json:"laugh"`
	Celebrate uint `json:"celebrate"`
	Fire      uint `json:"fire"`
	Support   uint `json:"support"`
	Sad       uint `json:"sad"`
	Eyes      uint `json:"eyes"`
}

type ReactionState struct {
	Reactions  ReactionCountsDTO `json:"reactions"`
	MyReaction *Reaction         `json:"myReaction"`
}

// 리액션 상태 변경에 필요한 파라미터 정의
type BoardReactionParam struct {
	BoardUid       uint
	PostUid        uint
	UserUid        uint
	Reaction       Reaction
	ReactionIsNull bool
	ReactionCode   ReactionType
	Notify         bool
	TargetUserUid  uint
}

type CommentReactionParam struct {
	BoardUid       uint
	CommentUid     uint
	UserUid        uint
	Reaction       Reaction
	ReactionIsNull bool
	ReactionCode   ReactionType
	Notify         bool
	TargetUserUid  uint
}
