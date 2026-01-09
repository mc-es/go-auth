package core

type Attr struct {
	Key   string
	Value any
}

func NewAttr(key string, value any) Attr {
	return Attr{
		Key:   key,
		Value: value,
	}
}

func SanitizeKey(k string) string {
	switch k {
	case "level", "msg", "time", "caller", "stacktrace":
		return "field." + k
	default:
		return k
	}
}
