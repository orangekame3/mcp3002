package mcp3002

import (
	"fmt"

	"golang.org/x/exp/io/spi"
)

type MCP3002 struct {
	Dev     *spi.Device
	Vref    float64
	Channel int
}

func (m MCP3002) Read() (float64, error) {
	if m.Channel < 0 || m.Channel > 1 { //    ---------------- (1)
		return 0, fmt.Errorf("channel is only selected 0 or 1")
	}
	cmd := byte(0x68) // ------------------------------------- (2)
	if m.Channel == 1 {
		cmd = 0x78
	}
	in := []byte{cmd, 0x00}
	out := make([]byte, 2)
	if err := m.Dev.Tx(in, out); err != nil {
		return 0, fmt.Errorf("failed to read channel %d,%w", m.Channel, err)
	}
	data := int(out[0]&3)<<8 | int(out[1]) //    ------------- (3)
	fmt.Printf("%b \n", out[0])
	fmt.Printf("%b \n", int(out[0]&3))
	fmt.Printf("%b \n", int(out[0]&3)<<8)
	return (m.Vref / 1024) * float64(data), nil
}
