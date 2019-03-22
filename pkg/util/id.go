package util

type IdGenerator struct {
	v int
}

func CreateIdGenerator(v int) *IdGenerator {
	return &IdGenerator{v: v}
}

func (g *IdGenerator) Next() int {
	g.v++
	return g.v
}
