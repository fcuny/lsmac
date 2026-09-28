package source

import "testing"

type countingCommand struct {
	calls int
}

func (c *countingCommand) Execute(binary string, args ...string) ([]byte, error) {
	c.calls++
	return []byte("output"), nil
}

func TestCachingCommandCachesByCommandLine(t *testing.T) {
	inner := &countingCommand{}
	cached := NewCachingCommand(inner)

	if _, err := cached.Execute("/usr/sbin/ioreg", "-rc", "IOPlatformExpertDevice", "-d1"); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if _, err := cached.Execute("/usr/sbin/ioreg", "-rc", "IOPlatformExpertDevice", "-d1"); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if inner.calls != 1 {
		t.Errorf("inner.calls = %d, want 1 (second call should be served from cache)", inner.calls)
	}

	if _, err := cached.Execute("/usr/sbin/ioreg", "-rc", "AGXAccelerator", "-d1"); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if inner.calls != 2 {
		t.Errorf("inner.calls = %d, want 2 (different command line must not be cached together)", inner.calls)
	}
}
