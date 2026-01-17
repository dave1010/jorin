package repl

// simple in-memory history ring buffer

type History interface {
	Add(line string)
	List(limit int) []string
}

type memHistory struct {
	cap  int
	buf  []string
	head int
	_    struct{}
}

func NewMemHistory(capacity int) History {
	if capacity <= 0 {
		capacity = 100
	}
	return &memHistory{cap: capacity, buf: make([]string, 0, capacity)}
}

func (h *memHistory) Add(line string) {
	if line == "" {
		return
	}
	if len(h.buf) < h.cap {
		h.buf = append(h.buf, line)
		return
	}
	h.buf[h.head] = line
	h.head = (h.head + 1) % h.cap
}

func (h *memHistory) List(limit int) []string {
	if limit <= 0 || limit > len(h.buf) {
		limit = len(h.buf)
	}
	res := make([]string, 0, limit)
	start := 0
	if len(h.buf) == h.cap {
		start = h.head
	}
	for i := 0; i < limit; i++ {
		idx := (start + i) % len(h.buf)
		res = append(res, h.buf[idx])
	}
	return res
}
