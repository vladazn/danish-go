package rand

import (
	"math/rand/v2"
	"time"
)

type Rand interface {
	IntN(n int) int
	Shuffle(n int, f func(i, j int))
}

type Random struct {
	r *rand.Rand
}

func New() *Random {
	return &Random{
		r: rand.New(rand.NewPCG(uint64(time.Now().Unix()/42), uint64(time.Now().Unix()/23))),
	}
}

func (r *Random) IntN(n int) int {
	return r.r.IntN(n)
}

func (r *Random) Shuffle(n int, f func(i, j int)) {
	r.r.Shuffle(n, f)
}
