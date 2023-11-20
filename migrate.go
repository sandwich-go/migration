package migration

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/sandwich-go/boost/xos"
	"github.com/sandwich-go/boost/xpanic"
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

	// MigrateOnly
	// "flask db migrate" only
	MigrateOnly(submitComment string) (err error)

	// ShowLocalRevision
	// Show the revision denoted by the given symbol.
	ShowLocalRevision(version string) (revision Revision, err error)

	// ShowDatabaseRevision
	// Shows the current revision of the database.
	ShowDatabaseRevision() (revision Revision, err error)

	// ShowDDL
	// Use The --sql option present in several commands performs an ‘offline’ mode migration.
	// Instead of executing the database commands the SQL statements that need to be executed are printed to the console.
	// params:
	// latest      - Write only the latest version of the update to the ddl file
	ShowDDL(latest bool) (ddl string, err error)

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
	Command(env string, name string, arg ...string) (output []byte, err error)
}

type migrate struct {
	logger *Logger
	conf   ConfInterface
}

func New(logger *log.Logger, opts ...ConfOption) Migration {
	return &migrate{logger: NewLogger(logger), conf: NewConf(opts...)}
}

func (g *migrate) flaskEnv() string {
	return fmt.Sprintf("FLASK_APP=%s", g.conf.GetFileName())
}

func (g *migrate) migrationBuildDir() (migrationBuildDir string) {
	return g.conf.GetScriptRoot()
}

func (g *migrate) Generate(opts ...GenerateConfOption) error {
	g.logger.Info("generate migration python script file...")
	conf := NewGenerateConf(opts...)
	args := []string{
		"migration",
		"--dir", g.migrationBuildDir(),
		"--file_name", g.conf.GetFileName(),
		"--db_host", conf.GetMysqlHost(),
		"--db_port", strconv.Itoa(conf.GetMysqlPort()),
		"--db_user", conf.GetMysqlUser(),
		"--db_pass", conf.GetMysqlPassword(),
		"--db_name", conf.GetMysqlDbName(),
		"--config", conf.GetProtokitGoSettingPath(),
		"--log_level=4",
	}

	var out []byte
	var err error
	xpanic.Try(func() {
		out, err = g.Command("", conf.GetProtokitPath(), args...)
	}).Catch(func(e xpanic.E) {
		err = fmt.Errorf("panic as error:%v", e)
	})
	g.logger.InfoWithFlag(err, "generate migration python script file", ", args:", HidePassword(args, conf.GetMysqlPassword()), ", out:", string(out))
	return err
}

func (g *migrate) Command(env string, name string, arg ...string) (output []byte, err error) {
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

func (g *migrate) prepare() (deferFunc func(), err error) {
	g.logger.Info("prepare...")
	var (
		output []byte
		dir    string
	)
	defer func() {
		g.logger.InfoWithFlag(err, "prepare", ", dir:", dir, ", file:", g.conf.GetFileName(), ", output:\n", string(output))
	}()
	dir = g.migrationBuildDir()
	deferFunc, err = Chdir(dir)
	if err != nil {
		return
	}
	output, err = g.Command(g.flaskEnv(), "flask", "db", "init")
	if err != nil {
		if strings.Contains(err.Error(), migrationsAlreadyExists) {
			g.logger.WarnWithFlag(migrationsAlreadyExists)
			err = nil
		} else if strings.Contains(err.Error(), migrationsAlreadyDone) {
			g.logger.WarnWithFlag(migrationsAlreadyDone)
			err = nil
		}
	}
	return
}

func (g *migrate) fetchDsnFromFile() (dsn string, err error) {
	g.logger.Info("fetch DSN from migration python script...")
	var migrationBuildDir string
	defer func() {
		g.logger.InfoWithFlag(err, "fetch DSN from migration python script", ", migrationBuildDir:", migrationBuildDir, ", file:", g.conf.GetFileName(), ", dsn:", dsn)
	}()

	migrationBuildDir = g.migrationBuildDir()
	file := filepath.Join(migrationBuildDir, g.conf.GetFileName())
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
	g.logger.Info("create database if not exists...")
	var dsn, dbName string
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

func (g *migrate) generateRevisionScript(_ string) (err error) {
	g.logger.Info("execute flask db migrate...")
	var output []byte
	defer func() {
		g.logger.InfoWithFlag(err, "execute flask db migrate", ", output:\n", string(output))
	}()

	// 检查是否有migrations/versions目录，versions目录为空的时候，git不会上传空目录
	// 需要手动创建一次 以免migrate报错
	// 检查目录是否存在
	dirPath := "./migrations/versions"
	_, err = os.Stat(dirPath)
	if os.IsNotExist(err) {
		err = os.Mkdir(dirPath, 0755)
		if err != nil {
			return
		}
	}

	// 用时"间戳+CommitID"作为本次migrate的提交内容(因为无法支持中文，且提交内容对用户无用)
	message := fmt.Sprintf(`--message=%d`, time.Now().Unix())
	// 用CommitID作为本次migrate的版本号
	revisionId := fmt.Sprintf(`--rev-id=%s`, g.conf.GetCommitID())

	output, err = g.Command(g.flaskEnv(), "flask", "db", "migrate", message, revisionId)
	if err != nil {
		if strings.Contains(err.Error(), dbNotUpToDate) {
			g.logger.WarnWithFlag(dbNotUpToDate)
			err = nil
		} else if strings.Contains(err.Error(), SchemaNoChanges) {
			g.logger.WarnWithFlag(SchemaNoChanges)
			err = nil
		}
	}
	return
}

func (g *migrate) Migrate(submitComment string) (revision Revision, err error) {
	var deferFunc func()
	deferFunc, err = g.prepare()
	defer deferFunc()
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

func (g *migrate) MigrateOnly(submitComment string) (err error) {
	var deferFunc func()
	deferFunc, err = g.prepare()
	defer deferFunc()
	if err != nil {
		return
	}
	return g.generateRevisionScript(submitComment)
}

type Revision struct {
	Rev        string
	Parent     string
	Path       string
	Message    string
	RevisionId string
	Revises    string
	CreateDate string
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
			source := strings.TrimSpace(strs[1])
			var target time.Time
			target, err = time.Parse("2006-01-02 15:04:05.999999", source)
			out[index].CreateDate = target.Format("2006-01-02 15:04:05")
		default:
			out[index].Message = strings.TrimSpace(strs[0])
		}
	}
	return out, err
}

func (g *migrate) ShowLocalRevision(version string) (revision Revision, err error) {
	g.logger.Info("show local revision...")
	var output []byte
	defer func() {
		g.logger.InfoWithFlag(err, "show local revision", ", revision:", revision, ", output:\n", string(output))
	}()
	var deferFunc func()
	deferFunc, err = g.prepare()
	defer deferFunc()
	if err != nil {
		return
	}
	if len(version) > 0 {
		output, err = g.Command(g.flaskEnv(), "flask", "db", "show", version)
	} else {
		output, err = g.Command(g.flaskEnv(), "flask", "db", "show")
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
	g.logger.Info("show remote revision...")
	var output []byte
	defer func() {
		g.logger.InfoWithFlag(err, "show remote revision", ", revision:", revision, ", output:\n", string(output))
	}()
	var deferFunc func()
	deferFunc, err = g.prepare()
	defer deferFunc()
	if err != nil {
		return
	}
	output, err = g.Command(g.flaskEnv(), "flask", "db", "current", "--verbose")
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

func (g *migrate) ShowDDL(latest bool) (ddl string, err error) {
	g.logger.Info("show ddl...")
	var output []byte
	defer func() {
		g.logger.InfoWithFlag(err, "show ddl", ", output:\n", string(output))
	}()
	var deferFunc func()
	deferFunc, err = g.prepare()
	defer deferFunc()
	if err != nil {
		return
	}
	output, err = g.Command(g.flaskEnv(), "flask", "db", "upgrade", "--sql")
	if err != nil {
		return
	}
	if latest {
		output, err = g.generateUpdateDDLFile(output)
		if err != nil {
			return "", err
		}
	}
	ddl = string(output)
	return
}

func (g *migrate) generateUpdateDDLFile(content []byte) (updateContent []byte, err error) {
	var deferFunc func()
	deferFunc, err = g.prepare()
	defer deferFunc()
	if err != nil {
		return
	}
	// 获取远程数据库的版本号
	var dbRevision, scriptRevision Revision
	if dbRevision, err = g.ShowDatabaseRevision(); err != nil {
		return
	}
	// 获取当前脚本的版本号
	if scriptRevision, err = g.ShowLocalRevision(""); err != nil {
		return
	}
	if scriptRevision.RevisionId == dbRevision.RevisionId {
		// 远程数据库最新，无需生成
		return
	}
	contents := strings.SplitN(string(content), scriptRevision.RevisionId, 2)
	if len(contents) > 1 {
		updateContent = []byte(contents[1])
	}
	return
}

func (g *migrate) deleteAlembicVersionUpdateAndInsertContent(content []byte) (ddlContent []byte, err error) {
	c := strings.TrimSpace(string(content))
	cs := strings.Split(c, ";")
	var contentAfterReplaced string
	for i := len(cs) - 1; i >= 0; i-- {
		checkContent := cs[i]
		if strings.Contains(checkContent, updateAlembicVersionPrefix) {
			contentAfterReplaced = strings.Replace(c, checkContent, "", -1)
		}
		if strings.Contains(checkContent, insertAlembicVersionPrefix) {
			contentAfterReplaced = strings.Replace(c, checkContent, "", -1)
		}
	}
	return []byte(contentAfterReplaced), nil
}

func (g *migrate) Upgrade() (err error) {
	g.logger.Info("upgrade...")
	var output []byte
	defer func() {
		g.logger.InfoWithFlag(err, "upgrade", ", output:\n", string(output))
	}()
	var deferFunc func()
	deferFunc, err = g.prepare()
	defer deferFunc()
	if err != nil {
		return
	}
	output, err = g.Command(g.flaskEnv(), "flask", "db", "upgrade")
	return
}

func (g *migrate) Downgrade() (err error) {
	g.logger.Info("downgrade...")
	var output []byte
	defer func() {
		g.logger.InfoWithFlag(err, "downgrade", ", output:\n", string(output))
	}()
	var deferFunc func()
	deferFunc, err = g.prepare()
	defer deferFunc()
	if err != nil {
		return
	}
	output, err = g.Command(g.flaskEnv(), "flask", "db", "downgrade")
	return
}

func (g *migrate) History() (revisions []Revision, err error) {
	g.logger.Info("history...")
	var output []byte
	defer func() {
		g.logger.InfoWithFlag(err, "history", ", output:\n", string(output))
	}()
	var deferFunc func()
	deferFunc, err = g.prepare()
	defer deferFunc()
	if err != nil {
		return
	}
	output, err = g.Command(g.flaskEnv(), "flask", "db", "history", "--verbose")
	if err != nil {
		return
	}
	return parseRevisions(string(output))
}
