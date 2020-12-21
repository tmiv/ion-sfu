package buffer

import (
	"encoding/binary"
	"sync"

	"github.com/pion/rtcp"
)

const maxPktSize = 1460

type Bucket struct {
	sync.Mutex
	buf    []byte
	nacker *nackQueue

	headSN   uint16
	step     int
	maxSteps int

	counter int
	onLost  func(nack []rtcp.NackPair)
}

func NewBucket(size int, nack bool) *Bucket {
	b := &Bucket{
		buf:      make([]byte, size),
		maxSteps: size/maxPktSize - 1,
	}
	if nack {
		b.nacker = newNACKQueue()
	}
	return b
}

func (b *Bucket) addPacket(pkt []byte, sn uint16, latest bool) []byte {
	b.Lock()
	defer b.Unlock()

	if !latest {
		b.nacker.remove(sn)
		return b.set(sn, pkt)
	}
	diff := sn - b.headSN
	b.headSN = sn
	for i := uint16(1); i < diff; i++ {
		b.step++
		if b.nacker != nil {
			b.counter++
			b.nacker.push(sn - i)
		}
		if b.step > b.maxSteps {
			b.step = 0
		}
	}

	if b.nacker != nil {
		b.counter++
		if b.counter > 2 {
			np := b.nacker.pairs()
			if len(np) > 0 {
				b.onLost(b.nacker.pairs())
			}
			b.counter = 0
		}
	}
	return b.push(pkt)
}

func (b *Bucket) getPacket(buf []byte, sn uint16) (i int, err error) {
	b.Lock()
	defer b.Unlock()
	p := b.get(sn)
	if p == nil {
		err = errPacketNotFound
		return
	}
	i = len(p)
	if len(buf) < i {
		err = errBufferTooSmall
		return
	}
	copy(buf, p)
	return
}

func (b *Bucket) push(pkt []byte) []byte {
	binary.BigEndian.PutUint16(b.buf[b.step*maxPktSize:], uint16(len(pkt)))
	off := b.step*maxPktSize + 2
	copy(b.buf[off:], pkt)
	b.step++
	if b.step > b.maxSteps {
		b.step = 0
	}
	return b.buf[off : off+len(pkt)]
}

func (b *Bucket) get(sn uint16) []byte {
	pos := b.step - int(b.headSN-sn+1)
	if pos < 0 {
		pos = b.maxSteps + pos + 1
	}
	off := pos * maxPktSize
	if off > len(b.buf) {
		return nil
	}
	if binary.BigEndian.Uint16(b.buf[off+4:off+6]) != sn {
		return nil
	}
	sz := int(binary.BigEndian.Uint16(b.buf[off : off+2]))
	return b.buf[off+2 : off+2+sz]
}

func (b *Bucket) set(sn uint16, pkt []byte) []byte {
	pos := b.step - int(b.headSN-sn+1)
	if pos < 0 {
		pos = b.maxSteps + pos + 1
	}
	off := pos * maxPktSize
	if off > len(b.buf) {
		return nil
	}
	binary.BigEndian.PutUint16(b.buf[off:], uint16(len(pkt)))
	copy(b.buf[off+2:], pkt)
	return b.buf[off+2 : off+2+len(pkt)]
}

func (b *Bucket) reset() {
	if b.headSN != 0 {
		b.headSN = 0
		b.counter = 0
		b.step = 0
		b.onLost = nil
		b.nacker.reset()
	}
}
