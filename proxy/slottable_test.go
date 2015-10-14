package proxy

import "testing"

func TestKey2Slot(t *testing.T) {
	pairs := map[string]string{
		"{user1000}.following": "user1000",
		"{user1000}.followers": "user1000",
		"foo{}{bar}":           "foo{}{bar}",
		"foo{{bar}}zap":        "{bar",
		"foo{bar}{zap}":        "bar",
		"{}bar":                "{}bar",
	}
	for k, v := range pairs {
		if Key2Slot(k) != int(CRC16([]byte(v))%NumSlots) {
			t.Errorf("slot not equal: %s, %s", k, v)
		}
	}
}
