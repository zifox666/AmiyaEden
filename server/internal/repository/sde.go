package repository

import (
	"fmt"
	"strings"
)

var sdeTranslationCategoryIDs = map[string]int{
	"type":          8,
	"group":         7,
	"category":      6,
	"description":   33,
	"tech":          34,
	"market_group":  36,
	"solar_system":  40,
	"constellation": 41,
	"region":        42,
}

// SdeNameMap 按 namespace 分组的名称映射。
type SdeNameMap map[string]map[int]string

// SdeRepository SDE 数据访问层
type SdeRepository struct{}

func NewSdeRepository() *SdeRepository { return &SdeRepository{} }

type sdeNaming struct {
	camelCase bool
}

func newSDENaming(camelCase bool) sdeNaming {
	return sdeNaming{camelCase: camelCase}
}

func (n sdeNaming) table(base string, alias ...string) string {
	name := base
	if !n.camelCase {
		name = strings.ToLower(base)
	}

	if n.camelCase {
		if len(alias) == 0 || alias[0] == "" {
			return fmt.Sprintf(`"%s"`, name)
		}
		return fmt.Sprintf(`"%s" %s`, name, alias[0])
	}

	if len(alias) == 0 || alias[0] == "" {
		return name
	}
	return fmt.Sprintf(`%s %s`, name, alias[0])
}

func (n sdeNaming) col(alias string, name string) string {
	if n.camelCase {
		return fmt.Sprintf(`%s."%s"`, alias, name)
	}
	return fmt.Sprintf(`%s.%s`, alias, strings.ToLower(name))
}

func (n sdeNaming) bareCol(name string) string {
	if n.camelCase {
		return fmt.Sprintf(`"%s"`, name)
	}
	return strings.ToLower(name)
}
