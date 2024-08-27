package utils

type void struct{}

// Compares two slices and returns slice of differences
// that not present in slice a 
func Missing(a, b []string) []string {
    // create map with length of the 'a' slice
    ma := make(map[string]void, len(a))
    diffs := []string{}
    // Convert first slice to map with empty struct (0 bytes)
    for _, ka := range a {
        ma[ka] = void{}
    }
    // find missing values in a
    for _, kb := range b {
        if _, ok := ma[kb]; !ok {
            diffs = append(diffs, kb)
        }
    }
    return diffs
}

func Contains(contents []string, item string) bool {
    for _, i := range contents {
        if item == i {
           return true 
        }
    }
    return false
}
