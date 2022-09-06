package rest

import "sort"

type Option int

const (
	AllowUnsuccessfulStatuses Option = iota
	FollowRedirects
	NoFollowRedirects
)

func HasOption(optionsList []Option, option Option) bool {
	sort.Slice(optionsList, func(i, j int) bool { return optionsList[i] < optionsList[j] })
	index := sort.Search(len(optionsList), func(i int) bool { return optionsList[i] == option })

	return index < len(optionsList)
}
