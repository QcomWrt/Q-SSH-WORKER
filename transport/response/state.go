package response

type State int

const (
    Continue State = iota // 301,302,307,308
    Connected             // 101 / 200
)