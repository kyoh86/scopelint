package scopelint

import (
	"strings"
)

const optionPrefix = "scopelint:"

func parseOptionComment(comment string) (options []string) {
	foreachOptionComment(comment, func(opt string) bool {
		options = append(options, opt)
		return true
	})
	return
}

func hasOptionComment(comment string, needle string) (having bool) {
	foreachOptionComment(comment, func(opt string) bool {
		if opt == needle {
			having = true
			return false
		}
		return true
	})
	return
}

func foreachOptionComment(comment string, walk func(option string) (_continue bool)) {
	for _, sentence := range strings.Split(comment, "//") {
		sentence = strings.TrimSpace(sentence)
		if !strings.HasPrefix(sentence, optionPrefix) {
			continue
		}
		sentence = strings.TrimSpace(strings.TrimPrefix(sentence, optionPrefix))
		for _, opt := range strings.Split(sentence, ",") {
			opt := strings.TrimSpace(opt)
			if opt != "" {
				if !walk(opt) {
					break
				}
			}
		}
	}
}
