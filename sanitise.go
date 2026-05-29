package main

import "strings"

func sanitiseChirp(chirp string) string {
	profanityVocabulary := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	split := strings.Split(chirp, " ")
	for idx, word := range split {
		if _, ok := profanityVocabulary[strings.ToLower(word)]; ok {
			split[idx] = strings.Repeat("*", 4)
		}
	}

	return strings.Join(split, " ")
}
