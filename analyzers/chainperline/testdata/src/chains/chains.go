// Eval для GID-196 (chainperline).
package chains

import (
	"strings"

	"github.com/sirupsen/logrus"

	"chains/sub"
)

// --- Позитив: цепочка из 2 вызовов в одну строку ---

func bad() string {
	return strings.NewReplacer("a", "b").Replace("aa") // want `GID-196: цепочка из 2 вызовов оформляется по одному вызову на строке, включая первый`
}

// --- Позитив: первый вызов на строке базы ---

func partial() string {
	return strings.NewReplacer("a", "b"). // want `GID-196: цепочка из 2 вызовов оформляется по одному вызову на строке, включая первый`
						Replace("aa")
}

// --- Позитив: цепочка через промежуточное поле ---

type job struct{}

func (j job) name() string { return "job" }

type repo struct{}

func (r repo) job() job { return job{} }

type svc struct{ r repo }

func fieldHop(s svc) string {
	return s.r.job().name() // want `GID-196: цепочка из 2 вызовов оформляется по одному вызову на строке, включая первый`
}

// --- Позитив: два звена на одной строке внутри многострочной цепочки ---

func twoOnOneLine(s svc) string {
	return s.r.
		job().name() // want `GID-196: цепочка из 2 вызовов оформляется по одному вызову на строке, включая первый`
}

// --- Негатив: каждый вызов на своей строке, включая первый ---

func good() string {
	return strings.
		NewReplacer("a", "b").
		Replace("aa")
}

// --- Негатив: одиночный вызов inline ---

func single() string {
	return strings.ToUpper("x")
}

// --- Граница: вложенный вызов — не цепочка, внутренняя цепочка ловится ---

func nested() string {
	return strings.ToUpper(strings.NewReplacer("a", "b").Replace("aa")) // want `GID-196: цепочка из 2 вызовов оформляется по одному вызову на строке, включая первый`
}

// --- Граница: конверсия через селектор — не звено ---

func conv(v string) string {
	return sub.Code(v).Upper()
}

// --- Граница: вызов на результате функции — функция считается базой ---

func factory() job { return job{} }

func fromFactory() string {
	return factory().name()
}

// --- Неприменимость: logrus-цепочка — зона GID-156 ---

func logIt(l *logrus.Logger) {
	l.WithField("a", 1).Info("x")
}
