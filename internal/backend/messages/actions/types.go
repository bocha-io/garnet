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

type Meteor struct {
	BasicAction
}

type Cover struct {
	BasicAction
}

type PiercingShot struct {
	BasicAction
}

type Sidestep struct {
	BasicAction
}

type DrainSword struct {
	BasicAction
}

type WhirlwindAxe struct {
	BasicAction
}

type AffectedCard struct {
	CardIDAttacked string `json:"cardidattacked"`
	PreviousHp     int64  `json:"previoushp"`
	CurrentHp      int64  `json:"currenthp"`
}

type MovedCard struct {
	CardID string `json:"id"`
	X      int64  `json:"x"`
	Y      int64  `json:"y"`
}

type AttackResponse struct {
	CardIDAttacked string `json:"cardidattacked"`
	PreviousHp     int64  `json:"previoushp"`
	CurrentHp      int64  `json:"currenthp"`
	X              int64  `json:"x"`
	Y              int64  `json:"y"`
	UUID           string `json:"uuid"`
	MsgType        string `json:"msgtype"`
	CardIDAttacker string `json:"cardidattacker"`
	Player         string `json:"player"`
	LeftOverMana   int64  `json:"leftovermana"`
}

type SkillResponse struct {
	AffectedCards  []AffectedCard `json:"affectedcards"`
	MovedCards     []MovedCard    `json:"movedcards"`
	UUID           string         `json:"uuid"`
	MsgType        string         `json:"msgtype"`
	CardIDAttacker string         `json:"cardidattacker"`
	Player         string         `json:"player"`
	LeftOverMana   int64          `json:"leftovermana"`
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

type Surrender struct {
	UUID    string `json:"uuid"`
	MsgType string `json:"msgtype"`
	MatchID string `json:"id"`
}

type SurrenderResponse struct {
	UUID          string `json:"uuid"`
	MsgType       string `json:"msgtype"`
	Loser         string `json:"loser"`
	LoserUsername string `json:"loserusername"`
}
