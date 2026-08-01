package tests

import "testing"

// TestRegExpExecUsesUTF16Indices verifies the externally visible index,
// lastIndex, capture, and has-indices values around an astral code point.
func TestRegExpExecUsesUTF16Indices(t *testing.T) {
	testSource(t, `
const match = /(😀)(a)/du.exec("😀a");
assertEqual(match.index, 0);
assertEqual(match[0], "😀a");
assertEqual(match[1], "😀");
assertEqual(match[2], "a");
assertEqual(match.indices[0][0], 0);
assertEqual(match.indices[0][1], 3);
assertEqual(match.indices[1][0], 0);
assertEqual(match.indices[1][1], 2);
assertEqual(match.indices[2][0], 2);
assertEqual(match.indices[2][1], 3);
assertEqual(match.indices.groups, undefined);

const global = /a/gu;
const globalMatch = global.exec("😀a");
assertEqual(globalMatch.index, 2);
assertEqual(global.lastIndex, 3);

const insidePair = /./gu;
insidePair.lastIndex = 1;
const pairMatch = insidePair.exec("😀a");
assertEqual(pairMatch.index, 0);
assertEqual(pairMatch[0], "😀");
assertEqual(insidePair.lastIndex, 2);

const failed = /z/gu;
failed.lastIndex = 1;
assertEqual(failed.exec("😀a"), null);
assertEqual(failed.lastIndex, 0);
`)
}

// TestRegExpNamedCaptureIndices verifies that only named captures create the
// null-prototype groups objects and that their ranges remain UTF-16 based.
func TestRegExpNamedCaptureIndices(t *testing.T) {
	testSource(t, `
const match = /(?<face>😀)(?<tail>a)/du.exec("😀a");
assertEqual(match.groups.face, "😀");
assertEqual(match.groups.tail, "a");
assertEqual(match.indices.groups.face[0], 0);
assertEqual(match.indices.groups.face[1], 2);
assertEqual(match.indices.groups.tail[0], 2);
assertEqual(match.indices.groups.tail[1], 3);

const unmatched = /(a)?b/d.exec("b");
assertEqual(unmatched[1], undefined);
assertEqual(unmatched.indices[1], undefined);

const flags = new RegExp("a", "dgimsuy");
assertEqual(flags.flags, "dgimsuy");
assertEqual(flags.toString(), "/a/dgimsuy");
assertEqual(new RegExp("a/b").toString(), "/a\\/b/");
`)
}
