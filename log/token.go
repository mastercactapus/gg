package log

// Token represents a parseable token from a log file.
type Token int

// Token types
const (
	TokenLineComment Token = iota
	TokenBlockComment
	TokenWord
	TokenLBrace
	TokenRBrace
	TokenComma
	TokenNumber
	TokenFlag
	TokenString
	TokenEquals
	TokenGT
	TokenLT
	TokenIdentifier
	TokenWhitespace
	TokenNewLine
	TokenIllegal
	TokenUnterminatedString
	TokenEOF // Returned when the reader reaches EOF
)
