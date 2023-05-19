package migration

const (
	exportFlaskApp             = "export FLASK_APP=%s"
	flaskDbInit                = "flask db init"
	flaskDbMigrate             = "flask db migrate"
	flaskDbUpgrade             = "flask db upgrade"
	flaskDbUpgradeDDL          = "flask db upgrade --sql"
	flaskDbDowngrade           = "flask db downgrade"
	flaskDbShow                = "flask db show"
	flaskDbCurrent             = "flask db current --verbose"
	flaskDbHistory             = "flask db history --verbose"
	migrationsAlreadyExists    = "Directory migrations already exists and is not empty"
	migrationsAlreadyDone      = "No handlers could be found for logger"
	dbNotUpToDate              = "Target database is not up to date"
	SchemaNoChanges            = "No changes in schema detected"
	mysqlDSNPrefix             = "mysql://"
	updateAlembicVersionPrefix = "UPDATE alembic_version"
	insertAlembicVersionPrefix = "INSERT INTO alembic_version"
)
