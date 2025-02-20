package set

func GetFirst(set map[string]struct{}) string {
	for e := range set {
		delete(set, e)
		return e
	}
	return ""
}
