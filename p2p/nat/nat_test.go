package nat

import (
	"bytes"
	"net"
	"testing"
	"time"
)

// This test checks that autodisc doesn't hang and returns
// consistent results when multiple goroutines call its methods
// concurrently.
func TestAutoDiscRace(t *testing.T) {
	ad := startautodisc("thing", func() Interface {
		time.Sleep(500 * time.Millisecond)
		return extIP{33, 44, 55, 66}
	})

	// Spawn a few concurrent calls to ad.ExternalIP.
	type rval struct {
		ip  net.IP
		err error
	}
	results := make(chan rval, 50)
	for i := 0; i < cap(results); i++ {
		go func() {
			ip, err := ad.ExternalIP()
			results <- rval{ip, err}
		}()
	}

	// Check that they all return the correct result within the deadline.
	deadline := time.After(2 * time.Second)
	for i := 0; i < cap(results); i++ {
		select {
		case <-deadline:
			t.Fatal("deadline exceeded")
		case rval := <-results:
			if rval.err != nil {
				t.Errorf("result %d: unexpected error: %v", i, rval.err)
			}
			wantIP := net.IP{33, 44, 55, 66}
			if !bytes.Equal(rval.ip, wantIP) {
				t.Errorf("result %d: got IP %v, want %v", i, rval.ip, wantIP)
			}
		}
	}
}

func TestAny(t *testing.T) {
	natm := Any()
	t.Logf("Auto Find NAT Type: %T", natm)
	t.Logf("Auto Find NAT Type: %T", natm.(*autodisc).found)

}

func TestMap(t *testing.T) {
	natm := Any()
	closing := make(chan struct{})
	go Map(natm, closing, "udp", 4501, 4501, "NAT-MAP")

	var realaddr net.UDPAddr

	if ext, err := natm.ExternalIP(); err == nil {
		realaddr = net.UDPAddr{IP: ext, Port: realaddr.Port} //获取在NAT网管上映射的公网地址
	}
	t.Logf("NAT Mapping Address : %v", realaddr.IP)
}

func TestStaticReflect(t *testing.T) {
	natm := StaticReflect("./ExIP.txt")
	closing := make(chan struct{})
	go Map(natm, closing, "udp", 4501, 4501, "NAT-MAP")

	var realaddr net.UDPAddr

	if ext, err := natm.ExternalIP(); err == nil {
		realaddr = net.UDPAddr{IP: ext, Port: realaddr.Port} //获取在NAT网管上映射的公网地址
	}
	t.Logf("NAT Mapping Address : %v", realaddr.IP)
}
