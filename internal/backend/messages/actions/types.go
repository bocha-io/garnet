package actions

const WorldID = "0x5FbDB2315678afecb367f032d93F642f64180aa3"

type BasicAction struct {
	UUID    string `json:"uuid"`
	MsgType string `json:"msgtype"`
	CardID  string `json:"id"`
	X       int64  `json:"x"`
	Y       int64  `json:"y"`
}

type PlaceCard struct {
	BasicAction
}

type PlaceCardResponse struct {
	UUID         string `json:"uuid"`
	MsgType      string `json:"msgtype"`
	CardID       string `json:"cardid"`
	X            int64  `json:"x"`
	Y            int64  `json:"y"`
	Player       string `json:"player"`
	LeftOverMana int64  `json:"leftovermana"`
}

type Attack struct {
	BasicAction
}

type AttackResponse struct {
	UUID           string `json:"uuid"`
	MsgType        string `json:"msgtype"`
	CardIDAttacker string `json:"cardidattacker"`
	CardIDAttacked string `json:"cardidattacked"`
	X              int64  `json:"x"`
	Y              int64  `json:"y"`
	PreviousHp     int64  `json:"previoushp"`
	CurrentHp      int64  `json:"currenthp"`
	Player         string `json:"player"`
	LeftOverMana   int64  `json:"leftovermana"`
}

type MoveCard struct {
	BasicAction
}

type MoveCardResponse struct {
	UUID         string `json:"uuid"`
	MsgType      string `json:"msgtype"`
	CardID       string `json:"cardid"`
	EndX         int64  `json:"endx"`
	EndY         int64  `json:"endy"`
	Player       string `json:"player"`
	LeftOverMana int64  `json:"leftovermana"`
}

type EndTurn struct {
	UUID    string `json:"uuid"`
	MsgType string `json:"msgtype"`
	MatchID string `json:"id"`
}

type EndTurnResponse struct {
	UUID    string `json:"uuid"`
	MsgType string `json:"msgtype"`
	Player  string `json:"player"`
	Mana    int64  `json:"mana"`
	Turn    int64  `json:"turn"`
}
