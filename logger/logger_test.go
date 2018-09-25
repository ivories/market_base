package logger

import "testing"

func TestLog(t *testing.T) {
	var std = New("testservice")
	std.LogDebug("aa")
	std.LogDebug("kafka", map[string]interface{}{})
	std.LogDebug("kafka", map[string]interface{}{"12": ""})
	std.LogInfo("kafka", map[string]interface{}{"12": ""})
}
