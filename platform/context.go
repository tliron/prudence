package platform

import (
	"strconv"
	"strings"
)

//
// VSTContext
//

type Context struct {
	Builder strings.Builder

	embedIndex int64
}

func (self *Context) NextSuffix() string {
	suffix := strconv.FormatInt(self.embedIndex, 10)
	self.embedIndex++
	return suffix
}

func (self *Context) WriteLiteral(literal string) {
	if literal != "" {
		self.Builder.WriteString("context.write('")
		for _, rune_ := range literal {
			switch rune_ {
			case '\n':
				self.Builder.WriteString("\\n")
			case '\'':
				self.Builder.WriteString("\\'")
			default:
				self.Builder.WriteRune(rune_)
			}
		}
		self.Builder.WriteString("');\n")
	}
}
