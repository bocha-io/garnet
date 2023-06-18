package data

type PlacedCards struct {
	P1cards int64 `json:"p1cards"`
	P2cards int64 `json:"p2cards"`
}

type Position struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

type CoverPosition struct {
	Card    string `json:"card"`
	Player  string `json:"player"`
	Card2   string `json:"card2"`
	Player2 string `json:"player2"`
	Raw     []Field
}

type MatchData struct {
	MatchID           string `json:"matchid"`
	PlayerOne         string `json:"playerone"`
	PlayerTwo         string `json:"playertwo"`
	PlayerOneUsermane string `json:"playeroneusername"`
	PlayerTwoUsermane string `json:"playertwousername"`
	CurrentTurn       int64  `json:"currenturn"`
	CurrentPlayer     string `json:"currentplayer"`
	CurrentMana       int64  `json:"currentmana"`
	Cards             []Card `json:"cards"`
}

type Card struct {
	ID            string   `json:"id"`
	Type          int64    `json:"type"`
	AttackDamage  int64    `json:"attackdamage"`
	MaxHp         int64    `json:"maxhp"`
	CurrentHp     int64    `json:"currenthp"`
	MovementSpeed int64    `json:"movementspeed"`
	Position      Position `json:"position"`
	Owner         string   `json:"owner"`
	ActionReady   bool     `json:"actionready"`
	Placed        bool     `json:"placed"`
}

type Base struct {
	ID        string `json:"id"`
	MaxHp     int64  `json:"maxhp"`
	CurrentHp int64  `json:"currenthp"`
}
