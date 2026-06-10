package server

import (
	_ "embed"
	"fmt"
)

//go:embed reloadScript.js
var liveReloadJS string

var liveReloadScript string

func init() {
	liveReloadScript = fmt.Sprintf(`
<script type="text/javascript">
(function() {
	if (!('EventSource' in window)) {
		console.error("Event Source is not found.");
		return;
	}
	%s
})();
</script>
		`, liveReloadJS)
}
