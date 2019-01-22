package mock

import (
	"log"
	"os"
	"strings"
	"testing"
)

func TestPanicWriter(t *testing.T) {
	log.SetOutput(&PanicWriter{})
	defer func() {
		log.SetOutput(os.Stderr)
		if r := recover(); r != nil {
			msg := r.(string)
			if !strings.HasSuffix(msg, "This is a panic\n") {
				t.Errorf("panic message is: %v", msg)
			}
		}
	}()
	log.Printf("This is a panic")
	t.Fail()
}
