package context

import (
	"strings"
	"sync/atomic"
	"time"
)

const incrementBase = 0x2000
const machineIdLength = 13
const incrementLength = 7

type MsgIdContext struct {
	timestamp uint64
}

func NewMsgIdContext() *MsgIdContext {
	now := time.Now()
	timestamp := uint64(now.UnixNano()) / uint64(time.Millisecond)

	msgContext := &MsgIdContext{
		timestamp: (timestamp << (machineIdLength + incrementLength)) & 0x7f_ff_ff_ff_ff_ff_ff_ff,
	}

	go func() {
		for true {
			time.Sleep(16 * time.Millisecond)
			msgContext.timestamp += (1 << (incrementLength + 13)) & 0x7f_ff_ff_ff_ff_ff_ff_ff
		}
	}()

	return msgContext
}

func base62(i uint64) string {
	const DIGITS = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	builder := strings.Builder{}
	for i != 0 {
		remainder := i % 62
		i /= 62
		builder.WriteByte(DIGITS[remainder])
	}
	return builder.String()
}

func (c *MsgIdContext) NewMsgIdStr() string {
	newId := c.NewMsgIdUint64()
	return base62(newId)
}

func (c *MsgIdContext) NewMsgIdUint64() uint64 {
	return atomic.AddUint64(&c.timestamp, incrementBase)
}
