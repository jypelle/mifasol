package tool

import "github.com/jypelle/mifasol/restApiV1"

func Deduplicate(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func DeduplicateArtistId(slice []restApiV1.ArtistId) []restApiV1.ArtistId {
	keys := make(map[restApiV1.ArtistId]bool)
	list := []restApiV1.ArtistId{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func DeduplicateUserId(slice []restApiV1.UserId) []restApiV1.UserId {
	keys := make(map[restApiV1.UserId]bool)
	list := []restApiV1.UserId{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
