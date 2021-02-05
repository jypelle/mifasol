package tool

func TernInt(exp bool, a int, b int) int {
	if exp {
		return a
	}
	return b
}

func TernStr(exp bool, a string, b string) string {
	if exp {
		return a
	}
	return b
}

func TernIface(exp bool, a interface{}, b interface{}) interface{} {
	if exp {
		return a
	}
	return b
}

func IfStr(exp bool, a string) string {
	if exp {
		return a
	}
	return ""
}

func IfMap(m map[string]interface{}, key string) bool {
	_, ok := m[key]

	return ok
}
