// Package egg holds the rotating footer logic shared by `ifly status`
// and the plugin's `/ifly:status` command pool. One in ten invocations
// shows the love footer; the rest rotate through neutral messages.
package egg

import "fmt"

var neutralPool = []string{"feeling lucky", "guards up", "stay sharp"}

// Footer selects one line given a rotation index (use $RANDOM % 10 or
// time.Now().UnixNano() % N at call sites). When easterEgg is true and
// index%10 == 0, returns the love footer.
func Footer(version string, easterEgg bool, index int) string {
	if easterEgg && index%10 == 0 {
		return "\U0001F49C ifly"
	}
	neutral := neutralPool[index%len(neutralPool)]
	return fmt.Sprintf("IFLy v%s — %s", version, neutral)
}
