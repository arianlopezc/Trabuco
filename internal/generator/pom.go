package generator

import (
	"github.com/arianlopezc/Trabuco/internal/config"
)

// generateParentPOM generates the parent pom.xml file
func (g *Generator) generateParentPOM() error {
	return g.writeTemplate("pom/parent.xml.tmpl", "pom.xml")
}

// generateModulePOM generates the pom.xml for a specific module
func (g *Generator) generateModulePOM(module string) error {
	var templateName string
	switch module {
	case config.ModuleModel:
		templateName = "pom/model.xml.tmpl"
	case config.ModuleJobs:
		templateName = "pom/jobs.xml.tmpl"
	case config.ModuleSQLDatastore:
		templateName = "pom/sqldatastore.xml.tmpl"
	case config.ModuleNoSQLDatastore:
		templateName = "pom/nosqldatastore.xml.tmpl"
	case config.ModuleShared:
		templateName = "pom/shared.xml.tmpl"
	case config.ModuleAPI:
		templateName = "pom/api.xml.tmpl"
	case config.ModuleWorker:
		templateName = "pom/worker.xml.tmpl"
	case config.ModuleEvents:
		templateName = "pom/events.xml.tmpl"
	case config.ModuleEventConsumer:
		templateName = "pom/eventconsumer.xml.tmpl"
	case config.ModuleMCP:
		templateName = "pom/mcp.xml.tmpl"
	default:
		return nil
	}

	outputPath := module + "/pom.xml"
	return g.writeTemplate(templateName, outputPath)
}
