package grbl

type rt byte

const (
	rtJogCancel   rt = 0x85
	rtSoftReset   rt = 0x18
	rtStatus      rt = '?'
	rtStartResume rt = '~'
	rtFeedHold    rt = '!'
)
