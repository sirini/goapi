package models

import "fmt"

// DB reaction_type 코드와 API 문자열 식별자를 고정한다.
const (
	REACTION_NONE     ReactionType = 0
	REACTION_LIKE     ReactionType = 1
	REACTION_BEST     ReactionType = 2
	REACTION_FACEPALM ReactionType = 3
	REACTION_HMM      ReactionType = 4
)

type ReactionType uint8

type Reaction = string

var REACTIONS = struct {
	LIKE     Reaction
	BEST     Reaction
	FACEPALM Reaction
	HMM      Reaction
}{LIKE: "like", BEST: "best", FACEPALM: "facepalm", HMM: "hmm"}

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
	default:
		return REACTION_NONE, fmt.Errorf("unknown reaction: %q", v)
	}
}

// 종류별 집계는 0이어도 항상 응답에 포함한다.
type ReactionCounts map[ReactionType]uint

type ReactionState struct {
	Reactions  ReactionCounts `json:"reactions"`
	MyReaction *Reaction      `json:"myReaction"`
}

// 리액션 상태 변경에 필요한 파라미터 정의
type BoardReactionParam struct {
	BoardUid      uint
	PostUid       uint
	UserUid       uint
	Reaction      Reaction
	ReactionCode  ReactionType
	Notify        bool
	TargetUserUid uint
}

type CommentReactionParam struct {
	BoardUid      uint
	CommentUid    uint
	UserUid       uint
	Reaction      Reaction
	ReactionCode  ReactionType
	Notify        bool
	TargetUserUid uint
}
