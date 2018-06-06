package rpcs

import (
	"strconv"
	"testing"
)

func TestB(t *testing.T) {

	rpc := NewRPCClient("PayoutsProcessor", "http://172.16.5.84:2801", "1200s")

	//gas, gasPrice, value string, autoGas bool
	for i := 0; i <= 300; i++ {
		a, err := rpc.SendTransactionParity("0x7684510bed649acb96fdde89c1d94f57d25512ce", "0x583cf945e967b213232ddc7019035f41608bf3e8", "100000", "1", strconv.Itoa(i*100), false, "123")
		t.Log(err)
		t.Log(a)
		t.Log(i)
	}
}
