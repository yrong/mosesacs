package cwmp

import (
	"testing"
)

func TestInformParsing(t *testing.T) {
	inform := `blabla`

	if inform != "blabla" {
		t.Errorf("Inform can't parse")
	}
}
