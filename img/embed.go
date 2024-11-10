package img

import (
	_ "embed"
)

var (
	//go:embed left-eye.png
	LeftEye []byte

	//go:embed right-eye.png
	RightEye []byte

	//go:embed nose-mouth.png
	NoseMouth []byte
)
