package geom

type Mat4Stack []Mat4

func (s *Mat4Stack) Push() {
	*s = append(*s, (*s)[len(*s)-1])
}

func (s *Mat4Stack) Pop() {
	*s = (*s)[:len(*s)-1]
}

func (s *Mat4Stack) Get() Mat4 {
	return (*s)[len(*s)-1]
}

func (s *Mat4Stack) GetPtr() *Mat4 {
	return &(*s)[len(*s)-1]
}

func (s *Mat4Stack) GetFloatPtr() *float32 {
	return &(*s)[len(*s)-1][0]
}

func (s *Mat4Stack) Load(m Mat4) {
	(*s)[len(*s)-1] = m
}

func (s *Mat4Stack) Multi(m Mat4) {
	(*s)[len(*s)-1] = (*s)[len(*s)-1].Mult(m)
}
