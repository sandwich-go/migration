package migration

import (
	"bitbucket.org/funplus/dbparser"
	"context"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/sandwich-go/protokit"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Migration interface {
	// Generate
	// Generate migration file and initializes migration support for the application.
	Generate(namespace *protokit.Namespace, dirs ...string) (err error)

	// Migrate
	// Create database if not exists.
	// Generate script revision and diff from remote database for update SQL DDL.
	Migrate(submitComment string) (revision Revision, err error)

	// ShowScriptRevision
	// Show the revision denoted by the given symbol.
	ShowScriptRevision(version string) (revision Revision, err error)

	// ShowDatabaseRevision
	// Shows the current revision of the database.
	ShowDatabaseRevision() (revision Revision, err error)

	// Upgrade
	// Upgrades the database.
	Upgrade() (err error)

	// Downgrade
	// Downgrades the database.
	Downgrade() (err error)

	// History
	// Shows the list of migrations.
	History() (revisions []Revision, err error)
}

type migrate struct {
	prepared bool
	cfg      *Config
}

func New(opts ...ConfigOption) Migration { return &migrate{cfg: NewConfig(opts...)} }

func (g *migrate) runCommand(name string, args ...string) (s string, err error) {
	return g.cfg.GetRunCommand()(name, args...)
}

func (g *migrate) runBashCommand(args ...string) (s string, err error) {
	return g.runCommand("bash", append([]string{"-c"}, args...)...)
}

func (g *migrate) getFileNameWithExt() string {
	return fmt.Sprintf("%s.py", g.cfg.GetFileName())
}

func (g *migrate) prepare() (err error) {
	if g.prepared {
		return
	}
	if _, err = g.runCommand("cd", g.cfg.GetScriptRoot()); err != nil {
		return err
	}
	if _, err = g.runBashCommand(fmt.Sprintf(exportFlaskApp, g.getFileNameWithExt())); err != nil {
		return err
	}
	if _, err = g.runBashCommand(flaskDbInit); err != nil {
		if !strings.Contains(err.Error(), migrationsAlreadyExists) {
			return
		}
		err = nil
	}
	if err == nil {
		g.prepared = true
	}
	return err
}

func (g *migrate) finish() error {
	g.prepared = false
	return nil
}

func (g *migrate) generateRevisionScript(submitComment string) error {
	var err error
	if err = g.prepare(); err != nil {
		return err
	}
	var message string
	if len(submitComment) > 0 {
		message = fmt.Sprintf(`--message="%s"`, submitComment)
	}
	if _, err = g.runBashCommand(flaskDbMigrate, message); err != nil {
		if !strings.Contains(err.Error(), dbNotUpToDate) {
			return err
		}
	}
	return nil
}

type Revision struct {
	Rev        string
	Parent     string
	Path       string
	Message    string
	RevisionId string
	Revises    string
	CreateDate time.Time
}

func parseRevisions(s string) ([]Revision, error) {
	var err error
	var valid bool
	var out []Revision
	var index = -1
	ss := strings.Split(s, "\n")
	for _, str := range ss {
		str = strings.TrimSpace(str)
		if len(str) == 0 {
			continue
		}
		strs := strings.SplitN(str, ":", 2)
		if strs[0] == "Rev" {
			valid = true
			index++
			out = append(out, Revision{})
		}
		if !valid {
			continue
		}
		switch strs[0] {
		case "Rev":
			out[index].Rev = strings.TrimSpace(strs[1])
		case "Parent":
			out[index].Parent = strings.TrimSpace(strs[1])
		case "Path":
			out[index].Path = strings.TrimSpace(strs[1])
		case "Revision ID":
			out[index].RevisionId = strings.TrimSpace(strs[1])
		case "Revises":
			out[index].Revises = strings.TrimSpace(strs[1])
		case "Create Date":
			out[index].CreateDate, err = time.Parse("2006-01-02 15:04:05", strings.TrimSpace(strs[1]))
		default:
			out[index].Message = strings.TrimSpace(strs[0])
		}
	}
	return out, err
}

func (g *migrate) ShowScriptRevision(version string) (revision Revision, err error) {
	if err = g.prepare(); err != nil {
		return
	}
	var showContent string
	if showContent, err = g.runBashCommand(flaskDbShow, version); err != nil {
		return
	}
	var revisions []Revision
	if revisions, err = parseRevisions(showContent); err != nil {
		return
	}
	if len(revisions) == 0 {
		return
	}
	return revisions[0], nil
}

func (g *migrate) ShowDatabaseRevision() (revision Revision, err error) {
	if err = g.prepare(); err != nil {
		return
	}
	var showContent string
	if showContent, err = g.runBashCommand(flaskDbCurrent); err != nil {
		return
	}
	var revisions []Revision
	if revisions, err = parseRevisions(showContent); err != nil {
		return
	}
	if len(revisions) == 0 {
		return
	}
	return revisions[0], nil
}

func (g *migrate) generateUpdateDDLFile() (err error) {
	if err = g.prepare(); err != nil {
		return
	}
	// 获取远程数据库的版本号
	var dbRevision, scriptRevision Revision
	if dbRevision, err = g.ShowDatabaseRevision(); err != nil {
		return
	}
	// 获取当前脚本的版本号
	if scriptRevision, err = g.ShowScriptRevision(""); err != nil {
		return
	}
	if scriptRevision.RevisionId == dbRevision.RevisionId {
		// 远程数据库最新，无需生成
		return
	}
	if len(dbRevision.RevisionId) == 0 {
		dbRevision.RevisionId = emptyRevision
	}
	var content string
	if content, err = g.runBashCommand(flaskDbUpgradeDDL); err != nil {
		return
	}
	contents := strings.SplitN(content, scriptRevision.RevisionId, 2)
	if len(contents) > 1 {
		err = FilePutContents(filepath.Join(g.cfg.GetScriptRoot(), fmt.Sprintf("%s_%s.sql", dbRevision.RevisionId, scriptRevision.RevisionId)), []byte(strings.TrimSpace(contents[1])))
	}
	return
}

func (g *migrate) Upgrade() (err error) {
	if err = g.prepare(); err != nil {
		return
	}
	if _, err = g.runBashCommand(flaskDbUpgrade); err != nil {
		return
	}
	return
}

func (g *migrate) Downgrade() (err error) {
	if err = g.prepare(); err != nil {
		return
	}
	if _, err = g.runBashCommand(flaskDbDowngrade); err != nil {
		return
	}
	return
}

func (g *migrate) Generate(namespace *protokit.Namespace, dirs ...string) error {
	parser := dbparser.New(namespace)
	if err := parser.Scanning(dirs...); err != nil {
		return err
	}
	content, err := parser.DumpMigrateContent(fmt.Sprintf("%s%s:%s@%s:%d/%s",
		mysqlDSNPrefix,
		g.cfg.GetMysqlUser(),
		g.cfg.GetMysqlPassword(),
		g.cfg.GetMysqlHost(),
		g.cfg.GetMysqlPort(),
		g.cfg.GetMysqlDbName(),
	))
	if err != nil {
		return err
	}
	return FilePutContents(filepath.Join(g.cfg.GetScriptRoot(), g.getFileNameWithExt()), []byte(content))
}

func (g *migrate) fetchDsnFromFile() (dsn string, err error) {
	file := g.getFileNameWithExt()
	// 需要解析下migration.py中的`SQLALCHEMY_DATABASE_URI`
	if !fileExists(file) {
		err = fmt.Errorf("not found '%s' migration file, need call 'GenerateMigration' function first", file)
		return
	}
	var content []byte
	content, err = FileGetContents(g.getFileNameWithExt())
	if err != nil {
		return
	}
	reg := regexp.MustCompile(`app.config\['SQLALCHEMY_DATABASE_URI'] = '\s*(.*)\s*'`)
	all := reg.FindAllStringSubmatch(string(content), -1)
	if len(all) < 1 || len(all[0]) < 2 {
		err = fmt.Errorf("invalid migration file, not found 'SQLALCHEMY_DATABASE_URI' in '%s'", file)
		return
	}
	// 需要查看dsn，是否有这样的库
	dsn = strings.TrimPrefix(all[0][1], mysqlDSNPrefix)
	dsns := strings.Split(dsn, ":")
	for i, j := range dsns {
		if strings.HasPrefix(j, "@") {
			dsns[i] = strings.Replace(j, "@", "@tcp(", 1)
			ss := strings.Split(dsns[i+1], "/")
			ss[0] += ")"
			dsns[i+1] = strings.Join(ss, "/")
			dsn = strings.Join(dsns, ":")
			break
		}
	}
	return
}

func (g *migrate) createDatabaseIfNotExists() error {
	dsn, err := g.fetchDsnFromFile()
	if err != nil {
		return err
	}
	var config *mysql.Config
	if config, err = mysql.ParseDSN(dsn); err != nil {
		return err
	}
	dbName := config.DBName
	config.DBName = ""
	var mdb *sql.DB
	if mdb, err = sql.Open("mysql", config.FormatDSN()); err != nil {
		return err
	}
	_, err = mdb.ExecContext(context.Background(), fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName))
	return err
}

func (g *migrate) Migrate(submitComment string) (revision Revision, err error) {
	if err = g.prepare(); err != nil {
		return
	}
	// 创建远程版本库
	if err = g.createDatabaseIfNotExists(); err != nil {
		return
	}
	if err = g.generateRevisionScript(submitComment); err != nil {
		return
	}
	if err = g.generateUpdateDDLFile(); err != nil {
		return
	}
	return g.ShowScriptRevision("")
}

func (g *migrate) History() (revisions []Revision, err error) {
	if err = g.prepare(); err != nil {
		return
	}
	var showContent string
	if showContent, err = g.runBashCommand(flaskDbHistory); err != nil {
		return
	}
	return parseRevisions(showContent)
}
