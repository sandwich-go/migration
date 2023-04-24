package migration

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	sh "github.com/codeskyblue/go-sh"
	"github.com/go-sql-driver/mysql"
	"github.com/sandwich-go/boost/xos"
	"github.com/sandwich-go/boost/xpanic"
	"github.com/sandwich-go/boost/xproc"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Migration interface {
	// Generate
	// Generate migration file and initializes migration support for the application.
	Generate(opts ...GenerateConfOption) (err error)

	// Migrate
	// Create database if not exists.
	// Generate script revision and diff from remote database for update SQL DDL.
	Migrate(submitComment string) (revision Revision, err error)

	// ShowLocalRevision
	// Show the revision denoted by the given symbol.
	ShowLocalRevision(version string) (revision Revision, err error)

	// ShowDatabaseRevision
	// Shows the current revision of the database.
	ShowDatabaseRevision() (revision Revision, err error)

	// ShowDDL
	ShowDDL(ddlFilePath string) (ddl string, err error)

	// Upgrade
	// Upgrades the database.
	Upgrade() (err error)

	// Downgrade
	// Downgrades the database.
	Downgrade() (err error)

	// History
	// Shows the list of migrations.
	History() (revisions []Revision, err error)

	// Command
	// Exec command.
	Command(name string, arg ...string) (output []byte, err error)
}

type migrate struct {
	logger *Logger
	conf   ConfInterface
}

func New(logger *log.Logger, opts ...ConfOption) Migration {
	return &migrate{logger: NewLogger(logger), conf: NewConf(opts...)}
}

func (g *migrate) Generate(opts ...GenerateConfOption) error {
	g.logger.Info("generate migration python script file...")
	var (
		err      error
		commitID string
	)

	out, gitErr := sh.Command("git", "show", "-s", "--format=%h").Output()
	if gitErr != nil {
		return fmt.Errorf("got err:%w while git show", gitErr)
	}

	commitID = strings.TrimSpace(string(out))
	if commitID == "" {
		return fmt.Errorf("got err: commitID is nil")
	}

	conf := NewGenerateConf(opts...)
	args := []string{
		"migration",
		"--dir", g.conf.GetScriptRoot(),
		"--file_name", g.conf.GetFileName(),
		"--db_host", conf.GetMysqlHost(),
		"--db_port", strconv.Itoa(conf.GetMysqlPort()),
		"--db_user", conf.GetMysqlUser(),
		"--db_pass", conf.GetMysqlPassword(),
		"--db_name", fmt.Sprintf("%s_%s", conf.GetMysqlDbName(), commitID),
		"--config", conf.GetProtokitGoSettingPath(),
		"--log_level=4",
	}

	xpanic.Try(func() {
		_, err = xproc.Run(conf.GetProtokitPath(), xproc.WithArgs(args...))
	}).Catch(func(err xpanic.E) {
		err = fmt.Errorf("panic as error:%v", err)
	})
	g.logger.InfoWithFlag(err, "generate migration python script file", ", args:", args)
	return err
}

func (g *migrate) Command(name string, arg ...string) (output []byte, err error) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	xpanic.Try(func() {
		cmd := exec.Command(name, arg...)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err = cmd.Run()
		output = stdout.Bytes()
	}).Catch(func(err xpanic.E) {
		err = fmt.Errorf("panic as error:%v", err)
	})
	if stderr.String() != "" {
		if err != nil {
			err = fmt.Errorf("error: %v, stderr:%s", err, stderr.String())
		}
	}
	return
}

func (g *migrate) CommandWithEnv(env string, name string, arg ...string) (output []byte, err error) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	xpanic.Try(func() {
		cmd := exec.Command(name, arg...)

		if env != "" {
			cmd.Env = append(cmd.Env, env)
		}

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err = cmd.Run()
		output = stdout.Bytes()
	}).Catch(func(err xpanic.E) {
		err = fmt.Errorf("panic as error:%v", err)
	})
	if stderr.String() != "" {
		if err != nil {
			err = fmt.Errorf("error: %v, stderr:%s", err, stderr.String())
		}
	}
	return
}

func (g *migrate) makeFlaskAppEnv() string {
	return fmt.Sprintf("FLASK_APP=%s", g.conf.GetFileName())
}

func (g *migrate) prepare() (err error) {
	g.logger.Info("prepare...")
	var output []byte
	defer func() {
		g.logger.InfoWithFlag(err, "prepare", ", script root:", g.conf.GetScriptRoot(), ", file:", g.conf.GetFileName(), ", output:\n", string(output))
	}()
	err = os.Chdir(g.conf.GetScriptRoot())
	if err != nil {
		return err
	}

	output, err = g.CommandWithEnv(g.makeFlaskAppEnv(), "flask", "db", "init")
	if err != nil {
		if strings.Contains(err.Error(), migrationsAlreadyExists) {
			g.logger.WarnWithFlag(migrationsAlreadyExists)
			err = nil
		} else if strings.Contains(err.Error(), migrationsAlreadyDone) {
			g.logger.WarnWithFlag(migrationsAlreadyDone)
			err = nil
		}
	}

	return err
}

func (g *migrate) fetchDsnFromFile() (dsn string, err error) {
	g.logger.Info("fetch DSN from migration python script...")
	defer func() {
		g.logger.InfoWithFlag(err, "fetch DSN from migration python script", ", script root:", g.conf.GetScriptRoot(), ", file:", g.conf.GetFileName(), ", dsn:", dsn)
	}()
	file := filepath.Join(g.conf.GetScriptRoot(), g.conf.GetFileName())
	// 需要解析下migration.py中的`SQLALCHEMY_DATABASE_URI`
	if !xos.ExistsFile(file) {
		err = fmt.Errorf("not found '%s' migration python script", file)
		return
	}
	var content []byte
	content, err = xos.FileGetContents(file)
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
		if strings.Contains(j, "@") {
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

func (g *migrate) createDatabaseIfNotExists() (err error) {
	var dsn, dbName string
	g.logger.Info("create database if not exists...")
	defer func() {
		g.logger.InfoWithFlag(err, "create database if not exists", ", dbName:", dbName)
	}()
	dsn, err = g.fetchDsnFromFile()
	if err != nil {
		return
	}
	var config *mysql.Config
	if config, err = mysql.ParseDSN(dsn); err != nil {
		return
	}
	dbName = config.DBName
	config.DBName = ""
	var mdb *sql.DB
	if mdb, err = sql.Open("mysql", config.FormatDSN()); err != nil {
		return err
	}
	_, err = mdb.ExecContext(context.Background(), fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName))
	return
}

func (g *migrate) generateRevisionScript(submitComment string) (err error) {
	var output []byte
	g.logger.Info("execute flask db migrate...")
	defer func() {
		g.logger.InfoWithFlag(err, "execute flask db migrate", ", output:\n", string(output))
	}()
	var message string
	if len(submitComment) > 0 {
		message = fmt.Sprintf(`--message="%s"`, submitComment)
	}
	output, err = g.CommandWithEnv(g.makeFlaskAppEnv(), "flask", "db", "migrate", message)
	if err != nil && strings.Contains(err.Error(), dbNotUpToDate) {
		err = nil
	}
	return
}

func (g *migrate) Migrate(submitComment string) (revision Revision, err error) {
	err = g.prepare()
	if err != nil {
		return
	}
	// 创建远程版本库
	err = g.createDatabaseIfNotExists()
	if err != nil {
		return
	}
	err = g.generateRevisionScript(submitComment)
	if err != nil {
		return
	}
	return g.ShowLocalRevision("")
}

type Revision struct {
	Rev        string
	Parent     string
	Path       string
	Message    string
	RevisionId string
	Revises    string
	CreateDate time.Time
	Content    string
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
			out = append(out, Revision{Content: s})
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

func (g *migrate) ShowLocalRevision(version string) (revision Revision, err error) {
	var output []byte
	g.logger.Info("show local revision...")
	defer func() {
		g.logger.InfoWithFlag(err, "show local revision", ", revision:", revision, ", output:\n", string(output))
	}()
	err = g.prepare()
	if err != nil {
		return
	}
	if len(version) > 0 {
		output, err = g.CommandWithEnv(g.makeFlaskAppEnv(), "flask", "db", "show", version)
	} else {
		output, err = g.CommandWithEnv(g.makeFlaskAppEnv(), "flask", "db", "show")
	}
	if err != nil {
		return
	}
	var revisions []Revision
	if revisions, err = parseRevisions(string(output)); err != nil {
		return
	}
	if len(revisions) == 0 {
		return
	}
	return revisions[0], nil
}

func (g *migrate) ShowDatabaseRevision() (revision Revision, err error) {
	var output []byte
	g.logger.Info("show remote revision...")
	defer func() {
		g.logger.InfoWithFlag(err, "show remote revision", ", revision:", revision, ", output:\n", string(output))
	}()
	err = g.prepare()
	if err != nil {
		return
	}
	output, err = g.CommandWithEnv(g.makeFlaskAppEnv(), "flask", "db", "current", "--verbose")
	if err != nil {
		return
	}
	var revisions []Revision
	if revisions, err = parseRevisions(string(output)); err != nil {
		return
	}
	if len(revisions) == 0 {
		return
	}
	return revisions[0], nil
}

func (g *migrate) ShowDDL(ddlFilePath string) (ddl string, err error) {
	var output []byte
	g.logger.Info("show ddl...")
	defer func() {
		g.logger.InfoWithFlag(err, "show ddl", ", output:\n", string(output))
	}()
	err = g.prepare()
	if err != nil {
		return
	}
	output, err = g.CommandWithEnv(g.makeFlaskAppEnv(), "flask", "db", "upgrade", "--sql")
	if err != nil {
		return
	}
	if len(ddlFilePath) > 0 {
		err = xos.FilePutContents(ddlFilePath, output)
	}
	ddl = string(output)
	return
}

//func (g *migrate) generateUpdateDDLFile() (err error) {
//	if err = g.prepare(); err != nil {
//		return
//	}
//	// 获取远程数据库的版本号
//	var dbRevision, scriptRevision Revision
//	if dbRevision, err = g.ShowDatabaseRevision(); err != nil {
//		return
//	}
//	// 获取当前脚本的版本号
//	if scriptRevision, err = g.ShowScriptRevision(""); err != nil {
//		return
//	}
//	if scriptRevision.RevisionId == dbRevision.RevisionId {
//		// 远程数据库最新，无需生成
//		return
//	}
//	if len(dbRevision.RevisionId) == 0 {
//		dbRevision.RevisionId = emptyRevision
//	}
//	var content string
//	if content, err = g.runBashCommand(flaskDbUpgradeDDL); err != nil {
//		return
//	}
//	contents := strings.SplitN(content, scriptRevision.RevisionId, 2)
//	if len(contents) > 1 {
//		err = FilePutContents(filepath.Join(g.cfg.GetScriptRoot(), fmt.Sprintf("%s_%s.sql", dbRevision.RevisionId, scriptRevision.RevisionId)), []byte(strings.TrimSpace(contents[1])))
//	}
//	return
//}

func (g *migrate) Upgrade() (err error) {
	var output []byte
	g.logger.Info("upgrade...")
	defer func() {
		g.logger.InfoWithFlag(err, "upgrade", ", output:\n", string(output))
	}()

	err = g.prepare()
	if err != nil {
		return
	}
	output, err = g.CommandWithEnv(g.makeFlaskAppEnv(), "flask", "db", "upgrade")
	return
}

func (g *migrate) Downgrade() (err error) {
	var output []byte
	g.logger.Info("downgrade...")
	defer func() {
		g.logger.InfoWithFlag(err, "downgrade", ", output:\n", string(output))
	}()

	err = g.prepare()
	if err != nil {
		return
	}
	output, err = g.CommandWithEnv(g.makeFlaskAppEnv(), "flask", "db", "downgrade")
	return
}

func (g *migrate) History() (revisions []Revision, err error) {
	var output []byte
	g.logger.Info("history...")
	defer func() {
		g.logger.InfoWithFlag(err, "history", ", output:\n", string(output))
	}()

	err = g.prepare()
	if err != nil {
		return
	}
	output, err = g.CommandWithEnv(g.makeFlaskAppEnv(), "flask", "db", "history", "--verbose")
	if err != nil {
		return
	}
	return parseRevisions(string(output))
}
