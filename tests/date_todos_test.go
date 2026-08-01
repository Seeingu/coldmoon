package tests

import "testing"

func TestDateStringMethodsUseECMAScriptShape(t *testing.T) {
	testSource(t, `
const epoch = new Date(0);
assertEqual(epoch.toString(), "Thu Jan 01 1970 00:00:00 GMT+0000");
assertEqual(epoch.toDateString(), "Thu Jan 01 1970");
assertEqual(epoch.toTimeString(), "00:00:00 GMT+0000");
assertEqual(epoch.getTimezoneOffset(), 0);
assertEqual(new Date(NaN).toString(), "Invalid Date");
`)
}
