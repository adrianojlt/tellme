package clipboard

import (
	"bytes"
	"fmt"
	"os/exec"
)

func Copy(text string) error {

	if path, err := exec.LookPath("pbcopy"); err == nil {
		
		cmd := exec.Command(path)
		cmd.Stdin = bytes.NewBufferString(text)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pbcopy: %w", err)
		}

		return nil
	}

	if path, err := exec.LookPath("xclip"); err == nil {

		cmd := exec.Command(path, "-selection", "clipboard")
		cmd.Stdin = bytes.NewBufferString(text)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("xclip: %w", err)
		}

		return nil
	}

	if path, err := exec.LookPath("xsel"); err == nil {

		cmd := exec.Command(path, "--clipboard", "--input")
		cmd.Stdin = bytes.NewBufferString(text)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("xsel: %w", err)
		}

		return nil
	}

	return fmt.Errorf("no clipboard tool found (tried pbcopy, xclip, xsel)")
}
