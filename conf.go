package migration

type RunCommand func(name string, arg ...string) (s string, err error)

//go:generate optiongen --option_with_struct_name=false --new_func=NewConfig --xconf=true --empty_composite_nil=true --usage_tag_name=usage
func ConfigOptionDeclareWithDefault() interface{} {
	return map[string]interface{}{
		"FileName":      "migration", // @MethodComment(migration 脚本名)
		"ScriptRoot":    ".",         // @MethodComment(migration 脚本根路径)
		"MysqlDbName":   "migration", // @MethodComment(migration db名)
		"MysqlUser":     "root",      // @MethodComment(migration 数据库用户名)
		"MysqlPassword": "",          // @MethodComment(migration 数据库用户密码)
		"MysqlHost":     "127.0.0.1", // @MethodComment(migration 数据库地址)
		"MysqlPort":     3306,        // @MethodComment(migration 数据库端口号)
		"RunCommand":    RunCommand(nil),
	}
}
