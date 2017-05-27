package log

// Token represents a parseable token from a log file.
type Token int

// Token types
const (
	TokenComment Token = iota
	TokenWord
	TokenSent
	TokenRecv
	TokenZero
	TokenLBrace
	TokenRBrace
	TokenComma
	TokenNumber
	TokenFlag
	TokenValue
	TokenIdentifier
	TokenWhitespace
	TokenNewLine
	TokenIllegal
	TokenEOF // Returned when the reader reaches EOF
)
