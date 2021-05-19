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

const base62Digits = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func base62(i uint64, builder *strings.Builder) *strings.Builder {
	if i == 0 {
		return builder
	}
	base62(i/62, builder)
	builder.WriteByte(base62Digits[i%62])
	return builder
}

func (c *MsgIdContext) NewMsgIdStr() string {
	newId := c.NewMsgIdUint64()
	return base62(newId, &strings.Builder{}).String()
}

func (c *MsgIdContext) NewMsgIdUint64() uint64 {
	return atomic.AddUint64(&c.timestamp, incrementBase)
}
