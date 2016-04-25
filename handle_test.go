package board_gamers

import (
	"reflect"
	"testing"
)

func TestExtractTrickplayGames(t *testing.T) {
	want := []string{"HAWAIIミニ拡張", "ロシアンレールロードミニ拡張＆ストーンエイジミニ拡張", "ヘックメック拡張"}
	text := "新しい神タイルや島タイルが含まれる「HAWAIIミニ拡張」、新しい技術者とボーナスタイルのセット「ロシアンレールロードミニ拡張＆ストーンエイジミニ拡張」、「ヘックメック拡張」が入荷しております。よろしくお願い致します。"
	if result := extractTrickplayGames(text); !reflect.DeepEqual(result, want) {
		t.Errorf("extractTrickplayGames = %v, want %v", result, want)
	}

	want = []string{"T.I.M.E Stories", "T.I.M.E Storiesシナリオ The Marcy Case"}
	text = "Space Cowboysが贈る壮大な謎解きゲーム「T.I.M.E Stories」、今度は異なる世界の過去の地球において、マーシィという女性を救う「T.I.M.E Storiesシナリオ The Marcy Case」 #トリックプレイ"
	if result := extractTrickplayGames(text); !reflect.DeepEqual(result, want) {
		t.Errorf("extractTrickplayGames = %v, want %v", result, want)
	}
}
