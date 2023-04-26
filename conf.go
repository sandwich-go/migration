package migration

//go:generate optiongen --option_with_struct_name=false --new_func=NewConf --xconf=true --empty_composite_nil=true --usage_tag_name=usage
func ConfOptionDeclareWithDefault() interface{} {
	return map[string]interface{}{
		"FileName":   "migration", // @MethodComment(migration 脚本名)
		"ScriptRoot": ".",         // @MethodComment(migration 脚本根路径)
		"CommitID":   "",          // @MethodComment(repo commitID)
		"DBName":     "",          // @MethodComment(migration 本地迁移库名)
	}
}

//go:generate optiongen --option_with_struct_name=false --new_func=NewGenerateConf --xconf=true --empty_composite_nil=true --usage_tag_name=usage
func GenerateConfOptionDeclareWithDefault() interface{} {
	return map[string]interface{}{
		"MysqlDbName":           "migration", // @MethodComment(migration db名)
		"MysqlUser":             "root",      // @MethodComment(migration 数据库用户名)
		"MysqlPassword":         "",          // @MethodComment(migration 数据库用户密码)
		"MysqlHost":             "127.0.0.1", // @MethodComment(migration 数据库地址)
		"MysqlPort":             3306,        // @MethodComment(migration 数据库端口号)
		"ProtokitGoSettingPath": "",          // @MethodComment(protokitgo 配置文件路径)
		"ProtokitPath":          "",          // @MethodComment(protokitgo 路径)
	}
}
