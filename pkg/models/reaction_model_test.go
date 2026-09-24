package models

import "testing"

// 새 종류까지 코드·문자열 왕복과 DTO 전체 키가 계약과 일치하는지 고정한다.
func TestReactionCodesRoundTrip(t *testing.T) {
	pairs := map[ReactionType]Reaction{
		REACTION_LIKE: REACTIONS.LIKE, REACTION_BEST: REACTIONS.BEST,
		REACTION_FACEPALM: REACTIONS.FACEPALM, REACTION_HMM: REACTIONS.HMM,
		REACTION_LAUGH: REACTIONS.LAUGH, REACTION_CELEBRATE: REACTIONS.CELEBRATE,
		REACTION_FIRE: REACTIONS.FIRE, REACTION_SUPPORT: REACTIONS.SUPPORT,
		REACTION_SAD: REACTIONS.SAD, REACTION_EYES: REACTIONS.EYES,
	}
	if len(ReactionCodeList) != len(pairs) {
		t.Fatalf("ReactionCodeList has %d entries, want %d", len(ReactionCodeList), len(pairs))
	}
	for code, value := range pairs {
		parsed, err := ParseReaction(value)
		if err != nil || parsed != code {
			t.Fatalf("ParseReaction(%q) = %d, %v; want %d", value, parsed, err, code)
		}
		if got := code.APIValue(); got != value {
			t.Fatalf("ReactionType(%d).APIValue() = %q, want %q", code, got, value)
		}
	}
	if _, err := ParseReaction("downvote"); err == nil {
		t.Fatal("ParseReaction should reject unknown reactions")
	}
	if REACTION_EYES != 10 {
		t.Fatalf("REACTION_EYES = %d, want 10", REACTION_EYES)
	}
}
