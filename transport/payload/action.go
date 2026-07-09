package payload

type ActionType byte

const (
    ActionWrite ActionType = iota
    ActionRead
)

type Action struct {
    Type ActionType
    Data string
}