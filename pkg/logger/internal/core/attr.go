package core

type Attr struct {
	Key   string
	Value any
}

func newAttr(k string, v any) Attr {
	return Attr{
		Key:   k,
		Value: v,
	}
}

func String(k, v string) Attr {
	return newAttr(k, v)
}

func Int(k string, v int) Attr {
	return newAttr(k, v)
}

func Bool(k string, v bool) Attr {
	return newAttr(k, v)
}

func Any(k string, v any) Attr {
	return newAttr(k, v)
}

func SanitizeKey(k string) string {
	switch k {
	case "level", "msg", "time", "caller", "stacktrace":
		return "field." + k
	default:
		return k
	}
}
