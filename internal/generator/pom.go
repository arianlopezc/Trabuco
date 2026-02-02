package generator

// generateParentPOM generates the parent pom.xml file
func (g *Generator) generateParentPOM() error {
	return g.writeTemplate("pom/parent.xml.tmpl", "pom.xml")
}

// generateModulePOM generates the pom.xml for a specific module
func (g *Generator) generateModulePOM(module string) error {
	var templateName string
	switch module {
	case "Model":
		templateName = "pom/model.xml.tmpl"
	case "SQLDatastore":
		templateName = "pom/sqldatastore.xml.tmpl"
	case "NoSQLDatastore":
		templateName = "pom/nosqldatastore.xml.tmpl"
	case "Shared":
		templateName = "pom/shared.xml.tmpl"
	case "API":
		templateName = "pom/api.xml.tmpl"
	case "Worker":
		templateName = "pom/worker.xml.tmpl"
	default:
		return nil
	}

	outputPath := module + "/pom.xml"
	return g.writeTemplate(templateName, outputPath)
}
