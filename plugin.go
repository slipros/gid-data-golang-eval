// Package gidrules регистрирует внутренние правила кода как линтеры
// golangci-lint через Module Plugin System.
//
// Каждое правило — отдельный линтер: его можно независимо включать и
// выключать в .golangci.yml (см. RULES.md — реестр правил).
package gidrules

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/analyzers/allptr"
	"github.com/slipros/gid-data-golang-eval/analyzers/bansymbol"
	"github.com/slipros/gid-data-golang-eval/analyzers/buildsig"
	"github.com/slipros/gid-data-golang-eval/analyzers/bytesinloop"
	"github.com/slipros/gid-data-golang-eval/analyzers/cacheplace"
	"github.com/slipros/gid-data-golang-eval/analyzers/chainperline"
	"github.com/slipros/gid-data-golang-eval/analyzers/chanbuf"
	"github.com/slipros/gid-data-golang-eval/analyzers/chandir"
	"github.com/slipros/gid-data-golang-eval/analyzers/constscope"
	"github.com/slipros/gid-data-golang-eval/analyzers/constvarorder"
	"github.com/slipros/gid-data-golang-eval/analyzers/convnaming"
	"github.com/slipros/gid-data-golang-eval/analyzers/createupdate"
	"github.com/slipros/gid-data-golang-eval/analyzers/ctornaming"
	"github.com/slipros/gid-data-golang-eval/analyzers/ctxkeys"
	"github.com/slipros/gid-data-golang-eval/analyzers/customctx"
	"github.com/slipros/gid-data-golang-eval/analyzers/dataresponse"
	"github.com/slipros/gid-data-golang-eval/analyzers/dbtags"
	"github.com/slipros/gid-data-golang-eval/analyzers/dirtree"
	"github.com/slipros/gid-data-golang-eval/analyzers/embedmutex"
	"github.com/slipros/gid-data-golang-eval/analyzers/entitygroup"
	"github.com/slipros/gid-data-golang-eval/analyzers/entitymethod"
	"github.com/slipros/gid-data-golang-eval/analyzers/enumconvert"
	"github.com/slipros/gid-data-golang-eval/analyzers/enumplace"
	"github.com/slipros/gid-data-golang-eval/analyzers/enumstring"
	"github.com/slipros/gid-data-golang-eval/analyzers/errlast"
	"github.com/slipros/gid-data-golang-eval/analyzers/errnew"
	"github.com/slipros/gid-data-golang-eval/analyzers/errplace"
	"github.com/slipros/gid-data-golang-eval/analyzers/errwrap"
	"github.com/slipros/gid-data-golang-eval/analyzers/eventctor"
	"github.com/slipros/gid-data-golang-eval/analyzers/exitonce"
	"github.com/slipros/gid-data-golang-eval/analyzers/failedto"
	"github.com/slipros/gid-data-golang-eval/analyzers/filterplace"
	"github.com/slipros/gid-data-golang-eval/analyzers/flagmain"
	"github.com/slipros/gid-data-golang-eval/analyzers/flatlayout"
	"github.com/slipros/gid-data-golang-eval/analyzers/fmtconst"
	"github.com/slipros/gid-data-golang-eval/analyzers/grpcinservice"
	"github.com/slipros/gid-data-golang-eval/analyzers/httperrors"
	"github.com/slipros/gid-data-golang-eval/analyzers/ifacemin"
	"github.com/slipros/gid-data-golang-eval/analyzers/ifacenaming"
	"github.com/slipros/gid-data-golang-eval/analyzers/ifaceplace"
	"github.com/slipros/gid-data-golang-eval/analyzers/initclean"
	"github.com/slipros/gid-data-golang-eval/analyzers/inlineconv"
	"github.com/slipros/gid-data-golang-eval/analyzers/inout"
	"github.com/slipros/gid-data-golang-eval/analyzers/intransaction"
	"github.com/slipros/gid-data-golang-eval/analyzers/layerimports"
	"github.com/slipros/gid-data-golang-eval/analyzers/logchain"
	"github.com/slipros/gid-data-golang-eval/analyzers/logconstruct"
	"github.com/slipros/gid-data-golang-eval/analyzers/logctx"
	"github.com/slipros/gid-data-golang-eval/analyzers/loggernew"
	"github.com/slipros/gid-data-golang-eval/analyzers/mapcap"
	"github.com/slipros/gid-data-golang-eval/analyzers/metricstruct"
	"github.com/slipros/gid-data-golang-eval/analyzers/modelmethod"
	"github.com/slipros/gid-data-golang-eval/analyzers/nilslice"
	"github.com/slipros/gid-data-golang-eval/analyzers/nobatch"
	"github.com/slipros/gid-data-golang-eval/analyzers/nogetprefix"
	"github.com/slipros/gid-data-golang-eval/analyzers/nopanic"
	"github.com/slipros/gid-data-golang-eval/analyzers/noptr"
	"github.com/slipros/gid-data-golang-eval/analyzers/onlygofrsuuid"
	"github.com/slipros/gid-data-golang-eval/analyzers/onlypkgerrors"
	"github.com/slipros/gid-data-golang-eval/analyzers/opstruct"
	"github.com/slipros/gid-data-golang-eval/analyzers/optsnaming"
	"github.com/slipros/gid-data-golang-eval/analyzers/optsstyle"
	"github.com/slipros/gid-data-golang-eval/analyzers/paramorder"
	"github.com/slipros/gid-data-golang-eval/analyzers/pkgstutter"
	"github.com/slipros/gid-data-golang-eval/analyzers/privatefunc"
	"github.com/slipros/gid-data-golang-eval/analyzers/receivernaming"
	"github.com/slipros/gid-data-golang-eval/analyzers/servicemodel"
	"github.com/slipros/gid-data-golang-eval/analyzers/servicesingle"
	"github.com/slipros/gid-data-golang-eval/analyzers/sqlnull"
	"github.com/slipros/gid-data-golang-eval/analyzers/subtestname"
	"github.com/slipros/gid-data-golang-eval/analyzers/upwardimport"
	"github.com/slipros/gid-data-golang-eval/analyzers/utilpkg"
	"github.com/slipros/gid-data-golang-eval/analyzers/validatorlib"
	"github.com/slipros/gid-data-golang-eval/analyzers/validatorshape"
)

//nolint:gochecknoinits // контракт golangci-lint Module Plugin System — регистрация только через init
func init() {
	register.Plugin("gidnogetprefix", newSingleAnalyzerPlugin(nogetprefix.Analyzer, register.LoadModeSyntax))
	register.Plugin("gidnobatch", newSingleAnalyzerPlugin(nobatch.Analyzer, register.LoadModeSyntax))
	register.Plugin("gidonlygofrsuuid", newSingleAnalyzerPlugin(onlygofrsuuid.Analyzer, register.LoadModeSyntax))
	register.Plugin("gidallptr", newSingleAnalyzerPlugin(allptr.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidflatlayout", newSingleAnalyzerPlugin(flatlayout.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("giddomainerrors", newSingleAnalyzerPlugin(errplace.DomainAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("giddalerrors", newSingleAnalyzerPlugin(errplace.DALAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidservicesingle", newSingleAnalyzerPlugin(servicesingle.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidonlypkgerrors", newSingleAnalyzerPlugin(onlypkgerrors.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidlayerimports", newConfigurablePlugin(layerimports.NewAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidservicemodel", newSingleAnalyzerPlugin(servicemodel.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidparamorder", newSingleAnalyzerPlugin(paramorder.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidoptsstyle", newSingleAnalyzerPlugin(optsstyle.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidlogconstruct", newSingleAnalyzerPlugin(logconstruct.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidlogctx", newSingleAnalyzerPlugin(logctx.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidlogchain", newSingleAnalyzerPlugin(logchain.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidconstvarorder", newSingleAnalyzerPlugin(constvarorder.Analyzer, register.LoadModeSyntax))
	register.Plugin("gidconstscope", newConfigurablePlugin(constscope.NewAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidmodelmethod", newConfigurablePlugin(modelmethod.NewAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidchainperline", newConfigurablePlugin(chainperline.NewAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidifacemin", newConfigurablePlugin(ifacemin.NewAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidinout", newConfigurablePlugin(inout.NewAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidcreateupdate", newConfigurablePlugin(createupdate.NewAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidprivatefunc", newSingleAnalyzerPlugin(privatefunc.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidentitygroup", newSingleAnalyzerPlugin(entitygroup.Analyzer, register.LoadModeSyntax))
	register.Plugin("giddirtree", newConfigurablePlugin(dirtree.NewAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidcacheplace", newConfigurablePlugin(cacheplace.NewAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidgrpcinservice", newConfigurablePlugin(grpcinservice.NewAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidnopanic", newSingleAnalyzerPlugin(nopanic.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidhttperrors", newSingleAnalyzerPlugin(httperrors.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("giddataresponse", newConfigurablePlugin(dataresponse.NewAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidvalidator", newConfigurablePlugin(validatorlib.NewAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidctxkeys", newSingleAnalyzerPlugin(ctxkeys.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidreceiver", newSingleAnalyzerPlugin(receivernaming.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidctor", newSingleAnalyzerPlugin(ctornaming.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidconvnaming", newSingleAnalyzerPlugin(convnaming.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidnoptr", newSingleAnalyzerPlugin(noptr.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidsqlnull", newSingleAnalyzerPlugin(sqlnull.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidenumstring", newSingleAnalyzerPlugin(enumstring.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("giddbtags", newConfigurablePlugin(dbtags.NewAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidenumbased", newSingleAnalyzerPlugin(enumstring.BasedAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidmodeltags", newConfigurablePlugin(dbtags.NewModelAnalyzer, register.LoadModeSyntax))
	register.Plugin("giderrfile", newConfigurablePlugin(errplace.NewFileAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidfilterplace", newSingleAnalyzerPlugin(filterplace.Analyzer, register.LoadModeSyntax))
	register.Plugin("gidifacenaming", newConfigurablePlugin(ifacenaming.NewAnalyzer, register.LoadModeSyntax))
	register.Plugin("gidmetricstruct", newSingleAnalyzerPlugin(metricstruct.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidintransaction", newSingleAnalyzerPlugin(intransaction.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("giderrwrap", newConfigurablePlugin(errwrap.NewWrapAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidstaticerr", newConfigurablePlugin(errwrap.NewStaticAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidifaceplace", newSingleAnalyzerPlugin(ifaceplace.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidembedmutex", newSingleAnalyzerPlugin(embedmutex.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidchanbuf", newSingleAnalyzerPlugin(chanbuf.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidinitclean", newConfigurablePlugin(initclean.NewAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidexitonce", newSingleAnalyzerPlugin(exitonce.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidbytesinloop", newSingleAnalyzerPlugin(bytesinloop.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidmapcap", newSingleAnalyzerPlugin(mapcap.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidfailedto", newConfigurablePlugin(failedto.NewAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidnilslice", newSingleAnalyzerPlugin(nilslice.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidfmtconst", newSingleAnalyzerPlugin(fmtconst.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidutilpkg", newConfigurablePlugin(utilpkg.NewAnalyzer, register.LoadModeSyntax))
	register.Plugin("gidcustomctx", newSingleAnalyzerPlugin(customctx.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidchandir", newSingleAnalyzerPlugin(chandir.Analyzer, register.LoadModeSyntax))
	register.Plugin("giderrlast", newSingleAnalyzerPlugin(errlast.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidsubtestname", newSingleAnalyzerPlugin(subtestname.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidflagmain", newSingleAnalyzerPlugin(flagmain.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidpkgstutter", newSingleAnalyzerPlugin(pkgstutter.Analyzer, register.LoadModeSyntax))
	register.Plugin("gidentitymethod", newConfigurablePlugin(entitymethod.NewAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidoptsnaming", newSingleAnalyzerPlugin(optsnaming.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidupwardimport", newSingleAnalyzerPlugin(upwardimport.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("giderrnew", newSingleAnalyzerPlugin(errnew.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidenumconvert", newSingleAnalyzerPlugin(enumconvert.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidopstruct", newSingleAnalyzerPlugin(opstruct.Analyzer, register.LoadModeSyntax))
	register.Plugin("gidenumplace", newSingleAnalyzerPlugin(enumplace.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidbuildsig", newSingleAnalyzerPlugin(buildsig.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidvalidatorshape", newConfigurablePlugin(validatorshape.NewAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidloggernew", newSingleAnalyzerPlugin(loggernew.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gidinlineconv", newSingleAnalyzerPlugin(inlineconv.Analyzer, register.LoadModeTypesInfo))
	register.Plugin("gideventctor", newConfigurablePlugin(eventctor.NewAnalyzer, register.LoadModeTypesInfo))
	register.Plugin("gidbansymbol", newConfigurablePlugin(bansymbol.NewAnalyzer, register.LoadModeTypesInfo))
}

// newSingleAnalyzerPlugin оборачивает один анализатор в плагин golangci-lint:
// одно правило = один линтер. loadMode — LoadModeSyntax для чисто
// AST-проверок, LoadModeTypesInfo для проверок с информацией о типах.
func newSingleAnalyzerPlugin(a *analysis.Analyzer, loadMode string) func(settings any) (register.LinterPlugin, error) {
	return func(_ any) (register.LinterPlugin, error) {
		return &singleAnalyzerPlugin{analyzer: a, loadMode: loadMode}, nil
	}
}

// newConfigurablePlugin — как newSingleAnalyzerPlugin, но анализатор
// строится из настроек линтера в .golangci.yml (settings).
func newConfigurablePlugin[T any](
	build func(T) *analysis.Analyzer,
	loadMode string,
) func(settings any) (register.LinterPlugin, error) {
	return func(settings any) (register.LinterPlugin, error) {
		cfg, err := register.DecodeSettings[T](settings)
		if err != nil {
			return nil, err
		}
		return &singleAnalyzerPlugin{analyzer: build(cfg), loadMode: loadMode}, nil
	}
}

type singleAnalyzerPlugin struct {
	analyzer *analysis.Analyzer
	loadMode string
}

func (s *singleAnalyzerPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{s.analyzer}, nil
}

//nolint:gidnogetprefix // имя метода — контракт интерфейса register.LinterPlugin
func (s *singleAnalyzerPlugin) GetLoadMode() string {
	return s.loadMode
}
