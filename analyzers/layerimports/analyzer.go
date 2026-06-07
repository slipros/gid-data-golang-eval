// Package layerimports реализует правила направления зависимостей между
// слоями Clean Architecture (полная матрица — RULES.md, GID-132/170/172,
// GID-224…229).
//
// GID-132 (layer-imports):
//   - /dal/** не импортирует /domain/** — repository работает только с entity;
//   - /domain/model не импортирует /dal/** — model чистый;
//   - /domain/usecase не импортирует /dal/** — usecase работает только
//     с model, с DAL общается через сервисы;
//   - /domain/service не импортирует /dal/repository — зависимость от
//     репозитория описывается интерфейсом рядом с потребителем.
//     Импорт /dal/entity сервису разрешён: он конвертирует model <-> entity.
//
// GID-170 (no-event-import):
//   - /domain/** не импортирует /event/**;
//   - /dal/** не импортирует /event/** — event-слой (kafka producer/consumer,
//     DTO) зависит от domain/model и конвертирует model <-> DTO, не наоборот.
//
// GID-172 (client-no-entity):
//   - /client/** не импортирует /dal/** — у клиента свои типы, он ничего
//     не знает о entity/repository.
//
// GID-224 (transport-imports):
//   - транспорт (/server, /schedule, /validate, /event) из слоёв сервиса
//     видит только /domain/model (и /validate) — конкретные service/usecase
//     инжектятся через интерфейсы у потребителя.
//
// GID-225 (root-and-leaves):
//   - /internal/app (composition root) и транспорт-листья (/server,
//     /schedule, /validate) никем не импортируются.
//
// GID-226 (metric-standalone):
//   - /metric не импортирует слои сервиса; domain/dal получают метрики
//     интерфейсом — /metric им недоступен (wiring в app).
//
// GID-227 (model-pure):
//   - /domain/model не импортирует ни один слой сервиса — это чистый
//     словарь; подпакеты /domain/model/* — полноправный model-слой.
//
// GID-228 (no-direct-client):
//   - /domain/** и /dal/** не импортируют /client/** — зависимость от
//     клиента описывается интерфейсом в /domain/model (GID-134).
//
// GID-229 (client-isolated):
//   - /client/** не импортирует слои сервиса (включая /domain целиком) —
//     у клиента свои типы, конвертация живёт у потребителя.
//
// Баны действуют только внутри одного модуля: для пакетов с сегментом
// /internal/ сравнивается префикс модуля, для остальных — первый сегмент
// пути (testdata, нестандартная раскладка). Сторонние библиотеки с
// сегментами client/event/metric в пути не задеваются.
//
// Ослабление per-project — settings.disable (список GID-ID); свои правила —
// settings.rules (id, scope, banned, reason). Точечно — //nolint:gidlayerimports.
package layerimports

import (
	"go/ast"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

// Порядок правил важен: для импорта рапортуется первое совпавшее правило
// (специфичные — раньше общих), дубли диагностик не плодятся.
var layerRules = []layerRule{
	{
		id:     "GID-132",
		scope:  []string{"dal"},
		banned: [][]string{{"domain"}},
		reason: "dal-слой работает только с entity, domain-типы ему недоступны",
	},
	{
		id:     "GID-132",
		scope:  []string{"domain", "model"},
		banned: [][]string{{"dal"}},
		reason: "model не зависит от dal-слоя",
	},
	{
		id:     "GID-132",
		scope:  []string{"domain", "usecase"},
		banned: [][]string{{"dal"}},
		reason: "usecase работает только с model, с DAL общается через сервисы",
	},
	{
		id:     "GID-132",
		scope:  []string{"domain", "service"},
		banned: [][]string{{"dal", "repository"}},
		reason: "сервис зависит от репозитория через интерфейс рядом с потребителем",
	},
	{
		id:     "GID-170",
		scope:  []string{"domain"},
		banned: [][]string{{"event"}},
		reason: "domain не зависит от event-слоя: event конвертирует model <-> DTO, не наоборот",
	},
	{
		id:     "GID-170",
		scope:  []string{"dal"},
		banned: [][]string{{"event"}},
		reason: "dal не зависит от event-слоя: event конвертирует model <-> DTO, не наоборот",
	},
	{
		id:     "GID-172",
		scope:  []string{"client"},
		banned: [][]string{{"dal"}},
		reason: "у клиента свои типы, он ничего не знает о entity/repository из dal-слоя",
	},
	// --- матрица изоляции слоёв (2026-06-07) ---
	{
		id:    "GID-227",
		scope: []string{"domain", "model"},
		banned: [][]string{
			{"domain", "service"}, {"domain", "usecase"},
			{"client"}, {"metric"}, {"server"}, {"schedule"}, {"validate"},
		},
		reason: "domain/model — чистый словарь сервиса, слои в него не текут",
	},
	{
		id:    "GID-226",
		scope: []string{"metric"},
		banned: [][]string{
			{"dal"}, {"domain"}, {"client"}, {"event"},
			{"server"}, {"schedule"}, {"validate"},
		},
		reason: "пакет metric — самостоятельный агрегатор Prometheus, слои сервиса ему недоступны",
	},
	{
		id:    "GID-229",
		scope: []string{"client"},
		banned: [][]string{
			{"domain"}, {"event"}, {"metric"},
			{"server"}, {"schedule"}, {"validate"},
		},
		reason: "у клиента свои типы: конвертация model <-> DTO клиента живёт у потребителя",
	},
	{
		id:    "GID-224",
		scope: []string{"server"},
		banned: [][]string{
			{"dal"}, {"domain", "service"}, {"domain", "usecase"},
			{"client"}, {"metric"}, {"event"}, {"schedule"}, {"app"},
		},
		reason: "транспорт работает только с domain/model: сервисы и зависимости инжектятся интерфейсами у потребителя",
	},
	{
		id:    "GID-224",
		scope: []string{"schedule"},
		banned: [][]string{
			{"dal"}, {"domain", "service"}, {"domain", "usecase"},
			{"client"}, {"metric"}, {"event"}, {"server"}, {"app"},
		},
		reason: "транспорт работает только с domain/model: сервисы и зависимости инжектятся интерфейсами у потребителя",
	},
	{
		id:    "GID-224",
		scope: []string{"validate"},
		banned: [][]string{
			{"dal"}, {"domain", "service"}, {"domain", "usecase"},
			{"client"}, {"metric"}, {"event"}, {"server"}, {"schedule"}, {"app"},
		},
		reason: "валидаторы работают только с domain/model и типами запросов",
	},
	{
		id:    "GID-224",
		scope: []string{"event"},
		banned: [][]string{
			{"dal"}, {"domain", "service"}, {"domain", "usecase"},
			{"client"}, {"metric"}, {"server"}, {"schedule"}, {"app"},
		},
		reason: "транспорт работает только с domain/model: сервисы и зависимости инжектятся интерфейсами у потребителя",
	},
	{
		id:     "GID-226",
		scope:  []string{"domain"},
		banned: [][]string{{"metric"}},
		reason: "domain получает метрики интерфейсом — пакет metric wiring'ует app",
	},
	{
		id:     "GID-226",
		scope:  []string{"dal"},
		banned: [][]string{{"metric"}},
		reason: "dal получает метрики интерфейсом — пакет metric wiring'ует app",
	},
	{
		id:     "GID-228",
		scope:  []string{"domain"},
		banned: [][]string{{"client"}},
		reason: "service/usecase зависят от клиента через интерфейс в domain/model, см. GID-134",
	},
	{
		id:     "GID-228",
		scope:  []string{"dal"},
		banned: [][]string{{"client"}},
		reason: "dal не вызывает внешние API напрямую — клиента wiring'ует app",
	},
	{
		id:     "GID-225",
		scope:  []string{"domain"},
		banned: [][]string{{"app"}, {"server"}, {"schedule"}, {"validate"}},
		reason: "composition root и транспорт — листья: их никто не импортирует",
	},
	{
		id:     "GID-225",
		scope:  []string{"dal"},
		banned: [][]string{{"app"}, {"server"}, {"schedule"}, {"validate"}},
		reason: "composition root и транспорт — листья: их никто не импортирует",
	},
	{
		id:     "GID-225",
		scope:  []string{"client"},
		banned: [][]string{{"app"}},
		reason: "composition root и транспорт — листья: их никто не импортирует",
	},
	{
		id:     "GID-225",
		scope:  []string{"metric"},
		banned: [][]string{{"app"}},
		reason: "composition root и транспорт — листья: их никто не импортирует",
	},
}

// Analyzer — правила GID-132/170/172/224…229 с настройками по умолчанию.
var Analyzer = NewAnalyzer(Settings{})

// Settings — настройки линтера из .golangci.yml.
type Settings struct {
	// Disable — GID-ID встроенных правил, которые проект осознанно отключает
	// (например, GID-224 на переходный период).
	Disable []string `json:"disable"`
	// Rules — дополнительные правила проекта поверх встроенной матрицы.
	Rules []RuleSetting `json:"rules"`
}

// RuleSetting — дополнительное правило направления импортов: пакетам
// слоя Scope запрещены импорты Banned. Слои задаются слэш-путями
// сегментов ("domain/service", "dal").
type RuleSetting struct {
	ID     string   `json:"id"`
	Scope  string   `json:"scope"`
	Banned []string `json:"banned"`
	Reason string   `json:"reason"`
}

// NewAnalyzer строит анализатор направления импортов между слоями.
func NewAnalyzer(s Settings) *analysis.Analyzer {
	rules := effectiveRules(s)
	return &analysis.Analyzer{
		Name: "gidlayerimports",
		Doc: "GID-132/GID-170/GID-172/GID-224…229: направление зависимостей " +
			"между слоями (матрица изоляции: dal/domain/server/schedule/" +
			"validate/event/client/metric/app)",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, rules)
		},
	}
}

// layerRule: пакетам в scope запрещены импорты banned. id — ID правила,
// под которым рапортуется нарушение.
type layerRule struct {
	id     string
	scope  []string
	banned [][]string
	reason string
}

// effectiveRules — встроенная матрица минус settings.disable плюс
// settings.rules (свои правила проверяются после встроенных).
func effectiveRules(s Settings) []layerRule {
	disabled := make(map[string]struct{}, len(s.Disable))
	for _, id := range s.Disable {
		disabled[id] = struct{}{}
	}
	rules := make([]layerRule, 0, len(layerRules)+len(s.Rules))
	//nolint:gidallptr // плагин не зависит от внутренней библиотеки gdhelper
	for _, rule := range layerRules {
		if _, ok := disabled[rule.id]; ok {
			continue
		}
		rules = append(rules, rule)
	}
	//nolint:gidallptr // плагин не зависит от внутренней библиотеки gdhelper
	for _, rs := range s.Rules {
		rule := layerRule{
			id:     rs.ID,
			scope:  segments(rs.Scope),
			reason: rs.Reason,
		}
		for _, b := range rs.Banned {
			if seg := segments(b); len(seg) > 0 {
				rule.banned = append(rule.banned, seg)
			}
		}
		if _, ok := disabled[rule.id]; ok || len(rule.scope) == 0 || len(rule.banned) == 0 {
			continue
		}
		rules = append(rules, rule)
	}
	return rules
}

// segments разбирает слэш-путь слоя ("domain/service") в сегменты.
func segments(path string) []string {
	var out []string
	for seg := range strings.SplitSeq(path, "/") {
		if seg != "" {
			out = append(out, seg)
		}
	}
	return out
}

func run(pass *analysis.Pass, rules []layerRule) (any, error) {
	pkgPath := pass.Pkg.Path()
	var scoped []layerRule
	//nolint:gidallptr // плагин не зависит от внутренней библиотеки gdhelper
	for _, rule := range rules {
		if pathseg.Contains(pkgPath, rule.scope...) {
			scoped = append(scoped, rule)
		}
	}
	if len(scoped) == 0 {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		checkImports(pass, scoped, file)
	}
	return nil, nil
}

func checkImports(pass *analysis.Pass, rules []layerRule, file *ast.File) {
	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		if !sameModule(pass.Pkg.Path(), path) {
			continue
		}
		reportFirstMatch(pass, rules, imp, path)
	}
}

// reportFirstMatch рапортует первое совпавшее правило: специфичные правила
// идут раньше общих, и один импорт не получает дубль диагностик.
func reportFirstMatch(pass *analysis.Pass, rules []layerRule, imp *ast.ImportSpec, path string) {
	//nolint:gidallptr // плагин не зависит от внутренней библиотеки gdhelper
	for _, rule := range rules {
		for _, banned := range rule.banned {
			if !pathseg.Contains(path, banned...) {
				continue
			}
			pass.Reportf(imp.Pos(),
				"%s: пакету %q запрещён импорт %q — %s",
				rule.id, pass.Pkg.Path(), path, rule.reason)
			return
		}
	}
}

// sameModule сообщает, принадлежит ли импорт тому же модулю, что и
// импортирующий пакет: слоевые баны не задевают сторонние библиотеки
// с сегментами client/event/metric в пути. Для канонической раскладки
// границей модуля служит сегмент /internal/; иначе (testdata,
// нестандартная раскладка) сравнивается первый сегмент пути.
func sameModule(pkgPath, importPath string) bool {
	const internalSeg = "/internal/"
	if module, _, ok := strings.Cut(pkgPath, internalSeg); ok {
		return strings.HasPrefix(importPath, module+internalSeg)
	}
	return firstSegment(pkgPath) == firstSegment(importPath)
}

func firstSegment(path string) string {
	seg, _, _ := strings.Cut(path, "/")
	return seg
}
