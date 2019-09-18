package bot

import "testing"

func Test_numre(t *testing.T) {
	parts := numRE.FindStringSubmatch("15")
	if parts == nil {
		t.Fail()
	}
}
