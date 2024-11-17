package typst

import "os/exec"

func Render(buf []byte) ([]byte, error) {

	cmd := exec.Command("typst", "Hello, World!")
}

func HasTypst()(bool,error){
	cmd := exec.Command("which", "typst")
	cmd.
}